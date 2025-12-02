mod api;

use crate::api::{ApiClient, ApiError, Board, UserStats};
use std::io::{self, Write};

const DEFAULT_BASE_URL: &str = "http://localhost:8321";

type CliResult<T> = Result<T, String>;

fn main() {
    if let Err(err) = run() {
        eprintln!("{err}");
        std::process::exit(1);
    }
}

fn run() -> CliResult<()> {
    println!("--- Minesweeper CLI ---");
    println!("Uses the API for gameplay; coordinates are 0-indexed.\n");

    let base_url = prompt_with_default("API base URL", DEFAULT_BASE_URL)?;
    let username = prompt_required("Username")?;
    let mut password = prompt_required("Password")?;

    let setup_client = ApiClient::new(&base_url);
    match setup_client.create_user(&username, &password) {
        Ok(_) => println!("New user created."),
        Err(ApiError::UnexpectedStatus(409, _)) => {}
        Err(err) => return Err(format!("Failed to create user: {err}")),
    }

    let client = loop {
        let client = ApiClient::with_credentials(&base_url, &username, &password);
        match client.get_user_stats() {
            Ok(stats) => {
                print_user_stats(&stats);
                break client;
            }
            Err(ApiError::UnexpectedStatus(401, _)) => {
                println!("Incorrect password. Please try again.");
                password = prompt_required("Password")?;
            }
            Err(err) => return Err(format!("Failed to find user: {err}")),
        }
    };

    loop {
        println!("Options:");
        println!("1) Start a new game");
        println!("2) See running games");
        println!("3) View my stats");
        println!("q) Quit");
        let choice = read_line("Select an option")?;
        match choice.as_str() {
            "1" => start_new_game(&client)?,
            "2" => resume_game(&client)?,
            "3" => show_user_stats(&client)?,
            "q" | "Q" => {
                println!("Goodbye!");
                return Ok(());
            }
            _ => println!("Please choose 1, 2, or q."),
        }
    }
}

fn start_new_game(client: &ApiClient) -> CliResult<()> {
    let width = prompt_i32("Board width", 10)?;
    let height = prompt_i32("Board height", 10)?;
    if width < 1 || height < 1 {
        println!("Width and height must both be at least 1.");
        return Ok(());
    }

    let max_bombs = width * height;
    let default_bombs = (max_bombs / 6).max(1);
    let bomb_prompt = format!("Bomb count (1-{max_bombs})");
    let bomb_count = loop {
        let bombs = prompt_i32(&bomb_prompt, default_bombs)?;
        if bombs >= 1 && bombs <= max_bombs {
            break bombs;
        }
        println!("Bombs must be between 1 and {max_bombs}.");
    };

    let response = client
        .new_game(width, height, bomb_count)
        .map_err(|err| format!("Failed to start game: {err}"))?;

    println!(
        "\nStarted game #{} ({}x{}, {} bombs). Enter moves as 'x y'.",
        response.id, width, height, bomb_count
    );
    play_game(client, response.id, response.board)
}

fn show_user_stats(client: &ApiClient) -> CliResult<()> {
    match client.get_user_stats() {
        Ok(stats) => {
            print_user_stats(&stats);
        }
        Err(err) => {
            return Err(format!("Failed to load user stats: {err}"));
        }
    }
    Ok(())
}

fn print_user_stats(stats: &UserStats) {
    println!(
        "\nLogged in as {} | Played: {} | Won: {} | Lost: {} | Avg moves: {:.2}\n",
        stats.username,
        stats.games_played,
        stats.games_won,
        stats.games_lost,
        stats.average_moves
    );
}

fn resume_game(client: &ApiClient) -> CliResult<()> {
    let games = client
        .get_unfinished_games()
        .map_err(|err| format!("Failed to fetch unfinished games: {err}"))?;

    if games.is_empty() {
        println!("No unfinished games found.");
        return Ok(());
    }

    println!("\nUnfinished games:");
    for (idx, game) in games.iter().enumerate() {
        let height = game.board.len();
        let width = game.board.first().map(|row| row.len()).unwrap_or(0);
        println!(
            "{}) Game #{} - {}x{} board, moves: {}, updated: {}",
            idx + 1,
            game.id,
            width,
            height,
            game.moves_count,
            game.updated_at
        );
    }

    let selection = loop {
        let choice = prompt_i32("Select a game by number", 1)?;
        if choice < 1 || (choice as usize) > games.len() {
            println!("Please choose a number between 1 and {}.", games.len());
            continue;
        }
        break (choice - 1) as usize;
    };

    let game = &games[selection];
    println!("\nResumed game #{}. Enter moves as 'x y'.", game.id);
    play_game(client, game.id, game.board.clone())
}

fn play_game(client: &ApiClient, game_id: i32, mut board: Board) -> CliResult<()> {
    loop {
        println!();
        render_board(&board);
        let input = read_line("Move (x y) or 'q' to return to menu")?;
        if input.eq_ignore_ascii_case("q") {
            println!("Returning to menu.");
            return Ok(());
        }

        let coords: Vec<_> = input.split_whitespace().collect();
        if coords.len() != 2 {
            println!("Enter a move as two numbers: x y");
            continue;
        }

        let x: i32 = match coords[0].parse() {
            Ok(val) => val,
            Err(_) => {
                println!("Could not read x coordinate.");
                continue;
            }
        };
        let y: i32 = match coords[1].parse() {
            Ok(val) => val,
            Err(_) => {
                println!("Could not read y coordinate.");
                continue;
            }
        };

        if board.is_empty() {
            println!("Board is empty; nothing to play.");
            return Ok(());
        }
        let height = board.len() as i32;
        let width = board[0].len() as i32;
        if x < 0 || y < 0 || x >= width || y >= height {
            println!("Coordinates must be within 0..{width} for x and 0..{height} for y.");
            continue;
        }

        let response = match client.make_move(game_id, x, y) {
            Ok(resp) => resp,
            Err(err) => {
                println!("Move failed: {err}");
                continue;
            }
        };
        board = response.board;

        if let Some(result) = response.result {
            render_board(&board);
            if result {
                println!("You win! Board cleared.");
            } else {
                println!("Boom! You hit a bomb.");
            }
            println!("Game #{game_id} finished.\n");
            return Ok(());
        }
    }
}

fn render_board(board: &Board) {
    if board.is_empty() {
        println!("(empty board)");
        return;
    }
    let width = board[0].len();

    print!("    ");
    for x in 0..width {
        print!("{:>3}", x);
    }
    println!();

    for (y, row) in board.iter().enumerate() {
        print!("{:>3} ", y);
        for cell in row {
            print!("{:>3}", display_cell(cell));
        }
        println!();
    }
}

fn display_cell(cell: &str) -> &str {
    match cell {
        " " => "#",
        "0" => " ",
        other => other,
    }
}

fn read_line(prompt: &str) -> CliResult<String> {
    print!("{prompt}: ");
    io::stdout().flush().map_err(|err| err.to_string())?;
    let mut input = String::new();
    io::stdin()
        .read_line(&mut input)
        .map_err(|err| err.to_string())?;
    Ok(input.trim().to_string())
}

fn prompt_with_default(prompt: &str, default: &str) -> CliResult<String> {
    let prompt_with_default = format!("{prompt} [{default}]");
    let input = read_line(&prompt_with_default)?;
    if input.is_empty() {
        Ok(default.to_string())
    } else {
        Ok(input)
    }
}

fn prompt_required(prompt: &str) -> CliResult<String> {
    loop {
        let input = read_line(prompt)?;
        if !input.is_empty() {
            return Ok(input);
        }
        println!("A value is required.");
    }
}

fn prompt_i32(prompt: &str, default: i32) -> CliResult<i32> {
    loop {
        let input = prompt_with_default(prompt, &default.to_string())?;
        match input.parse::<i32>() {
            Ok(val) => return Ok(val),
            Err(_) => println!("Please enter a valid number."),
        }
    }
}
