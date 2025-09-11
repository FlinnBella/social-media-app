



export type User = {
    sessionId: string;
    apiKey: string;
    username: string;
}


export type UserAuth = User & {
    JWTToken: string;
}