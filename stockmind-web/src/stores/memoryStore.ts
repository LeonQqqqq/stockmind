import { create } from 'zustand';
import type { Experience, Opinion } from '../types';

interface MemoryState {
  experiences: Experience[];
  opinions: Opinion[];
  searchKeyword: string;

  setExperiences: (experiences: Experience[]) => void;
  setOpinions: (opinions: Opinion[]) => void;
  setSearchKeyword: (keyword: string) => void;
  addExperience: (experience: Experience) => void;
  removeExperience: (id: number) => void;
  updateExperience: (experience: Experience) => void;
  removeOpinion: (id: number) => void;
}

export const useMemoryStore = create<MemoryState>((set) => ({
  experiences: [],
  opinions: [],
  searchKeyword: '',

  setExperiences: (experiences) => set({ experiences }),
  setOpinions: (opinions) => set({ opinions }),
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
  removeOpinion: (id) => set((state) => ({
    opinions: state.opinions.filter((o) => o.id !== id),
  })),
}));
