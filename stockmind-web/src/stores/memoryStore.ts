import { create } from 'zustand';
import type { Experience } from '../types';

interface MemoryState {
  experiences: Experience[];
  searchKeyword: string;

  setExperiences: (experiences: Experience[]) => void;
  setSearchKeyword: (keyword: string) => void;
  addExperience: (experience: Experience) => void;
  removeExperience: (id: number) => void;
  updateExperience: (experience: Experience) => void;
}

export const useMemoryStore = create<MemoryState>((set) => ({
  experiences: [],
  searchKeyword: '',

  setExperiences: (experiences) => set({ experiences }),
  setSearchKeyword: (keyword) => set({ searchKeyword: keyword }),
  addExperience: (experience) => set((state) => ({
    experiences: [experience, ...state.experiences],
  })),
  removeExperience: (id) => set((state) => ({
    experiences: state.experiences.filter((e) => e.id !== id),
  })),
  updateExperience: (experience) => set((state) => ({
    experiences: state.experiences.map((e) => e.id === experience.id ? experience : e),
  })),
}));
