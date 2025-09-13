



export type UserAuth = {
    sessionId: string;
    apiKey?: string;
    hasJWTToken: boolean;
    userSubscription: 'free' | 'pro';
}


export type User = UserAuth & {
    username: string;
    email: string;
}