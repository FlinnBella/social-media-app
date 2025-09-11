import {UserAuth, User} from '#types/user';
import { createContext } from 'react';

type AuthContextType= {
    isAuthenticated: boolean;
    setIsAuthenticated: (isAuthenticated: boolean) => void;
    Authenticate: (User: UserAuth,) => void;
    isAuthorized: (User: UserAuth) => boolean;
    Logout: () => void;


    // JWT actions
    generateJWTToken: (userAuth: UserAuth) => void;
    verifyJWTToken: (userAuth: UserAuth) => boolean;
    setJWTToken: (user: User) => void;
}



export const AuthContext = createContext<AuthContextType | undefined>(undefined);

/*
TODO: Define what AuthContext will look like. 
*/