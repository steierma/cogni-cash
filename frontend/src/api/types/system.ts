export interface SystemInfo {
    storage_mode: string;
    db_host: string;
    db_state: string;
    version: string;
    bank_provider?: string;
}

export interface User {
    id: string;
    username: string;
    email: string;
    full_name: string;
    address: string;
    role: string;
}

export interface AuthResponse {
    token: string;
    refresh_token: string;
}

export interface BridgeAccessToken {
    id: string;
    user_id: string;
    name: string;
    last_used_at?: string;
    created_at: string;
}

export interface CreateBridgeTokenResponse {
    token: string;
    info: BridgeAccessToken;
}