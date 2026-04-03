import React, { useEffect } from "react";

const AuthContext = React.createContext();

export function AuthProvider({ children }) {
    const [user, setUser] = React.useState(null);
    const [loading, setLoading] = React.useState(true);

    useEffect(() => {
        fetch(`${process.env.REACT_APP_AUTH_URL}/auth/me`, {
            method: 'GET',
            credentials: 'include',
        })
        .then(res => {
            if (res.ok) return res.json();
            return null;
        })
        .then(data => {
            setUser(data);
            setLoading(false);
        })
        .catch(() => {
            setUser(null);
            setLoading(false);
        });
    }, []);

    return (
        <AuthContext.Provider value={{ user, setUser, loading }}>
            { children } 
        </AuthContext.Provider>
    );
}

export function useAuth() {
    return React.useContext(AuthContext);
}