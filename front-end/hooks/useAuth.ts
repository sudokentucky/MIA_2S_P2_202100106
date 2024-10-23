import {create} from "zustand";

interface AuthState {
    isLogged: boolean;
    Login:()=>void;
    Logout:()=>void;
}

export const useAuth = create<AuthState>((set) => ({
    isLogged: false,
    Login: () => set({ isLogged: true }),
    Logout: () => set({ isLogged: false }),
}));