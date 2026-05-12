import { create } from 'zustand';

interface UIStore {
  showChat: boolean;
  animationsEnabled: boolean;
  soundEnabled: boolean;
  toggleChat: () => void;
  toggleAnimations: () => void;
  toggleSound: () => void;
}

export const useUIStore = create<UIStore>((set) => ({
  showChat: false,
  animationsEnabled: true,
  soundEnabled: true,
  toggleChat: () => set((s) => ({ showChat: !s.showChat })),
  toggleAnimations: () =>
    set((s) => ({ animationsEnabled: !s.animationsEnabled })),
  toggleSound: () => set((s) => ({ soundEnabled: !s.soundEnabled })),
}));
