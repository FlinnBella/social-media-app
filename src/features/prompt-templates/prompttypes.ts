export type PromptCards = PromptCard[];

export type PromptCard = {
    header: string; 
    description: string;
    image: string;
    promptType: keyof typeof PROMPT_TYPES;
}


export const PROMPT_TYPES = {
    REALTOR: 'REALTOR',
    JEWELER: 'JEWELER',
    EQUIPMENT_DEALER: 'EQUIPMENT_DEALER',
}

export const PromptCards = [
    {
        header: 'Realtor Imagery',
        description: 'Generate imagery for a realtor',
        image: '/house_image.webp',
        alt: 'Car Image',
        promptType: PROMPT_TYPES.REALTOR,
    },
    {
        header: 'Jeweler Imagery',
        description: 'Generate imagery for a jeweler',
        image: '/watch_image.webp',
        alt: 'Watch Image',
        promptType: PROMPT_TYPES.JEWELER,
    },
    {
        header: 'Equipment Dealer Imagery',
        description: 'Generate imagery for a equipment dealer',
        image: '/car_image.webp',
        alt: 'Car Image',
        promptType: PROMPT_TYPES.EQUIPMENT_DEALER,
    }
]