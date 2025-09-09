import { toast } from "sonner";
import { useMultiPartFormData, MULTIPART_ACTIONS } from "@/hooks/useMultiPartFormData";
import type {  FinalVideoResponse } from "#types/multipart";
import { API_ENDPOINTS, type ApiEndpointKey } from "@/cfg";

export interface SubmitVideoRequestArgs {
  prompt: string;
  images: File[];
  apiKey: ApiEndpointKey;
}

export async function submitVideoRequest(
  args: SubmitVideoRequestArgs
): Promise< FinalVideoResponse | Error> {
  const { prompt, images, apiKey } = args;

  if (!prompt || !prompt.trim()) {
    const message = "Please enter a property description";
    toast.error(message);
    return { name: 'No prompt entered', message: message } as Error;
  }
  if (!images || images.length === 0) {
    const message = "Please upload at least one image";
    toast.error(message);
    return { name: 'No images uploaded', message: message } as Error;
  }

  const formData = new FormData();
  formData.append("prompt", prompt);
  for (const file of images) {
    formData.append("image", file, file.name);
  }

  const path = API_ENDPOINTS[apiKey];
  const url = import.meta.env.PROD ? path : `http://localhost:8080${path}`;

  const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.SendImageTimeline, url);
  return resp as FinalVideoResponse;
}


