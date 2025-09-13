import {UserAuth} from '#types/user';
import { createContext, useEffect, useState } from 'react';
import { ApiEndpointKey } from '@/cfg';
import { jose } from "jose";

type AuthContextType= {
    setUserAuth: (userAuth: UserAuth) => void;
    UserAuth: UserAuth;
    isAuthenticated: boolean;
    setIsAuthenticated: (isAuthenticated: boolean) => void;
    Authenticate: (User: UserAuth,) => void;
    isAuthorized: (User: UserAuth, api_endpoint: ApiEndpointKey) => boolean;


    // JWT actions
    generateJWTToken: (userAuth: UserAuth) => UserAuth;
    hasJWTToken: (userAuth: UserAuth) => boolean;
    getJWTToken: (checkForJWT : ((userAuth: UserAuth) => boolean)) => any | null | Error;
}



export const AuthContext = createContext<AuthContextType | undefined>(undefined);


export const AuthProvider: React.FC<React.PropsWithChildren> = ({children}) => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [userHasJWTToken, setUserHasJWTToken] = useState(false);
    const [userAuth, setUserAuth] = useState<UserAuth | null>(null);

    useEffect(() => {
        if (hasJWTToken(userAuth as UserAuth)) {
            setUserHasJWTToken(true);
            return;
        } else {
            setUserAuth(generateJWTToken(userAuth as UserAuth) as UserAuth);
            return; 
        }

    }, [userAuth]);

    const Authenticate = (userAuth: UserAuth) => {
        setIsAuthenticated(true);
        setUserAuth(userAuth);
    }

    const isAuthorized = (userAuth: UserAuth, api_endpoint: ApiEndpointKey) => {
        return userAuth.apiKey === api_endpoint;
    }

    const generateJWTToken = (userAuth: UserAuth) => {
        
        //impl

        return userAuth;
    }

    const hasJWTToken = (userAuth: UserAuth) => {
        userAuth.hasJWTToken = userHasJWTToken;
        return userAuth.hasJWTToken;
    }

    const getJWTToken = (checkForJWT : ((userAuth: UserAuth) => boolean)) => {
        return userAuth;
    }

    return (
        <AuthContext.Provider value={{isAuthenticated, setIsAuthenticated, Authenticate, isAuthorized, generateJWTToken, hasJWTToken, getJWTToken, setUserAuth, UserAuth: userAuth as UserAuth }}>{children}</AuthContext.Provider>
    )
}