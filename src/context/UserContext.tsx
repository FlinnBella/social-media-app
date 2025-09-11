import { createContext, useContext, useState, PropsWithChildren } from "react";
import {
  useImageSelection,
} from "@/utilites/useImageSelection";

type UserStateContextType = {
  prompt: string;
  setPrompt: (next: string) => void;
  selectedFiles: File[];
  previewUrls: string[];
  handleFileSelect: (event: React.ChangeEvent<HTMLInputElement>) => void;
  removeSelectedFile: (index: number) => void;
  clearAllSelected: () => void;
  fileInputRef: React.RefObject<HTMLInputElement>;
  cameraInputRef: React.RefObject<HTMLInputElement>;
};

const UserStateContext = createContext<UserStateContextType | undefined>(undefined);

export function UserStateProvider({ children }: PropsWithChildren) {
  const [prompt, setPrompt] = useState("");

  const {
    selectedFiles,
    previewUrls,
    handleFileSelect,
    removeSelectedFile,
    clearAllSelected,
    fileInputRef,
    cameraInputRef,
  } = useImageSelection();

  const value: UserStateContextType = {
    prompt,
    setPrompt,
    selectedFiles,
    previewUrls,
    handleFileSelect,
    removeSelectedFile,
    clearAllSelected,
    fileInputRef,
    cameraInputRef,
  };

  return <UserStateContext.Provider value={value}>{children}</UserStateContext.Provider>;
}

export function useUserState(): UserStateContextType {
  const ctx = useContext(UserStateContext);
  if (!ctx) {
    throw new Error("useUserState must be used within a UserStateProvider");
  }
  return ctx;
}


