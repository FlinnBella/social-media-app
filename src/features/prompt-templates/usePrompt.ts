import { PROMPT_TYPES } from './prompttypes';


export const usePrompt = (prompt: string, promptType: keyof typeof PROMPT_TYPES) => {

    let ctx = "";

    switch (promptType) {
        case PROMPT_TYPES.REALTOR:
            ctx = "You are a realtor, you are given a prompt and a list of images, you need to generate a video about the property";
            return `${prompt} + ${ctx}`;
        case PROMPT_TYPES.JEWELER:
            ctx = "You are a jeweler, you are given a prompt and a list of images, you need to generate a video about the property";
            return `${prompt} + ${ctx}`;
        case PROMPT_TYPES.EQUIPMENT_DEALER:
            ctx = "You are a equipment dealer, you are given a prompt and a list of images, you need to generate a video about the property";
            return `${prompt} + ${ctx}`;
        default:
            throw new Error("Invalid prompt type");

    }
}