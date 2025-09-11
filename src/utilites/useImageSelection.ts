import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

export const ALLOWED_IMAGE_MIME_TYPES = new Set<string>([
  "image/jpeg",
  "image/png",
  "image/jpg",
]);

export const ALLOWED_IMAGE_EXTENSIONS = new Set<string>([".jpeg", ".jpg", ".png"]);

export const ALLOWED_FORMATS_LABEL = "JPEG (.jpg, .jpeg), PNG (.png)";

export const IMAGE_ACCEPT_ATTRIBUTE = ".jpg,.jpeg,.png,image/jpeg,image/png";

export function isAllowedImageFile(file: File): boolean {
  const type = file.type?.toLowerCase?.() || "";
  if (type && ALLOWED_IMAGE_MIME_TYPES.has(type)) return true;
  const name = file.name?.toLowerCase?.() || "";
  for (const ext of ALLOWED_IMAGE_EXTENSIONS) {
    if (name.endsWith(ext)) return true;
  }
  return false;
}

export function useImageSelection() {
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [previewUrls, setPreviewUrls] = useState<string[]>([]);
  const previewUrlsRef = useRef<string[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    previewUrlsRef.current.forEach((url) => {
      try {
        URL.revokeObjectURL(url);
      } catch {}
    });
    const urls = selectedFiles.map((file) => URL.createObjectURL(file));
    previewUrlsRef.current = urls;
    setPreviewUrls(urls);
    return () => {
      urls.forEach((url) => {
        try {
          URL.revokeObjectURL(url);
        } catch {}
      });
    };
  }, [selectedFiles]);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files) return;
    const allFiles = Array.from(files);
    const allAllowed = allFiles.every(isAllowedImageFile);
    if (!allAllowed) {
      toast.error(`Only these image formats are allowed: ${ALLOWED_FORMATS_LABEL}`);
      return;
    }
    setSelectedFiles((prev) => [...prev, ...allFiles]);
  };

  const removeSelectedFile = (index: number) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const clearAllSelected = () => {
    setSelectedFiles([]);
  };

  return {
    selectedFiles,
    setSelectedFiles,
    previewUrls,
    handleFileSelect,
    removeSelectedFile,
    clearAllSelected,
    fileInputRef,
    cameraInputRef,
  } as const;
}


