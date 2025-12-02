use reqwest::Method;
use reqwest::blocking::{Client, Response};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use std::fmt;

pub type Board = Vec<Vec<String>>;

#[derive(Debug, Clone)]
pub struct Credentials {
    username: String,
    password: String,
}

#[derive(Debug, Clone)]
pub struct ApiClient {
    base_url: String,
    http: Client,
    credentials: Option<Credentials>,
}

#[derive(Debug)]
pub enum ApiError {
    Http(reqwest::Error),
    MissingCredentials,
    UnexpectedStatus(u16, String),
}

impl fmt::Display for ApiError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ApiError::Http(err) => write!(f, "HTTP error: {err}"),
            ApiError::MissingCredentials => {
                write!(f, "Missing credentials for authenticated request")
            }
            ApiError::UnexpectedStatus(code, body) => {
                if body.is_empty() {
                    write!(f, "Unexpected status {code}")
                } else {
                    write!(f, "Unexpected status {code}: {body}")
                }
            }
        }
    }
}

impl std::error::Error for ApiError {}

impl From<reqwest::Error> for ApiError {
    fn from(value: reqwest::Error) -> Self {
        ApiError::Http(value)
    }
}

pub type ApiResult<T> = Result<T, ApiError>;

impl ApiClient {
    /// Create a client without credentials; use this for unauthenticated calls like user creation.
    pub fn new(base_url: impl Into<String>) -> Self {
        let trimmed = base_url.into().trim_end_matches('/').to_string();
        Self {
            base_url: trimmed,
            http: Client::new(),
            credentials: None,
        }
    }

    /// Create a client with Basic Auth credentials for protected endpoints.
    pub fn with_credentials(
        base_url: impl Into<String>,
        username: impl Into<String>,
        password: impl Into<String>,
    ) -> Self {
        let trimmed = base_url.into().trim_end_matches('/').to_string();
        Self {
            base_url: trimmed,
            http: Client::new(),
            credentials: Some(Credentials {
                username: username.into(),
                password: password.into(),
            }),
        }
    }

    pub fn create_user(&self, username: &str, password: &str) -> ApiResult<()> {
        let url = format!("{}/user", self.base_url);
        let resp = self
            .http
            .put(url)
            .json(&CreateUserRequest { username, password })
            .send()?;
        handle_empty(resp)
    }

    pub fn get_user_stats(&self) -> ApiResult<UserStats> {
        let resp = self.authed_request(Method::GET, "/user")?.send()?;
        parse_json(resp)
    }

    pub fn new_game(&self, width: i32, height: i32, bomb_count: i32) -> ApiResult<NewGameResponse> {
        let resp = self
            .authed_request(Method::PUT, "/game")?
            .json(&NewGameRequest {
                width,
                height,
                bomb_count,
            })
            .send()?;
        parse_json(resp)
    }

    pub fn make_move(&self, game_id: i32, x: i32, y: i32) -> ApiResult<MakeMoveResponse> {
        let resp = self
            .authed_request(Method::POST, "/game")?
            .json(&MakeMoveRequest { game_id, x, y })
            .send()?;
        parse_json(resp)
    }

    pub fn get_unfinished_games(&self) -> ApiResult<Vec<UnfinishedGame>> {
        let resp = self.authed_request(Method::GET, "/games")?.send()?;
        parse_json(resp)
    }

    fn authed_request(
        &self,
        method: Method,
        path: &str,
    ) -> ApiResult<reqwest::blocking::RequestBuilder> {
        let creds = self
            .credentials
            .as_ref()
            .ok_or(ApiError::MissingCredentials)?;
        let url = format!("{}{}", self.base_url, path);
        Ok(self
            .http
            .request(method, url)
            .basic_auth(&creds.username, Some(&creds.password)))
    }
}

fn handle_empty(resp: Response) -> ApiResult<()> {
    if resp.status().is_success() {
        return Ok(());
    }
    let status = resp.status().as_u16();
    let body = resp.text().unwrap_or_default();
    Err(ApiError::UnexpectedStatus(status, body))
}

fn parse_json<T: DeserializeOwned>(resp: Response) -> ApiResult<T> {
    if resp.status().is_success() {
        return Ok(resp.json()?);
    }
    let status = resp.status().as_u16();
    let body = resp.text().unwrap_or_default();
    Err(ApiError::UnexpectedStatus(status, body))
}

#[derive(Serialize)]
struct CreateUserRequest<'a> {
    username: &'a str,
    password: &'a str,
}

#[derive(Serialize)]
struct NewGameRequest {
    width: i32,
    height: i32,
    #[serde(rename = "bombCount")]
    bomb_count: i32,
}

#[derive(Serialize)]
struct MakeMoveRequest {
    #[serde(rename = "gameId")]
    game_id: i32,
    x: i32,
    y: i32,
}

#[derive(Debug, Deserialize)]
pub struct UserStats {
    pub username: String,
    #[serde(rename = "gamesPlayed")]
    pub games_played: i32,
    #[serde(rename = "gamesWon")]
    pub games_won: i32,
    #[serde(rename = "gamesLost")]
    pub games_lost: i32,
    #[serde(rename = "averageMoves")]
    pub average_moves: f64,
}

#[derive(Debug, Deserialize)]
pub struct NewGameResponse {
    pub id: i32,
    pub board: Board,
}

#[derive(Debug, Deserialize)]
pub struct UnfinishedGame {
    pub id: i32,
    pub board: Board,
    #[serde(rename = "movesCount")]
    pub moves_count: i32,
    #[serde(rename = "updatedAt")]
    pub updated_at: String,
}

#[derive(Debug, Deserialize)]
pub struct MakeMoveResponse {
    pub board: Board,
    pub result: Option<bool>,
}
