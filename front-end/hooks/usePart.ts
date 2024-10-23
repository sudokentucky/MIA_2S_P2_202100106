import {create} from "zustand";

interface PartState {
    isMounted: boolean;
    mount:()=>void;
    unmount:()=>void;
}

export const usePart = create<PartState>((set) => ({
    isMounted: false,
    mount: () => set({ isMounted: true }),
    unmount: () => set({ isMounted: false }),
}));

