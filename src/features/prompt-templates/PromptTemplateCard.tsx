import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/base-components/Card';
import { PROMPT_TYPES } from './prompttypes';
import { PromptCard } from './prompttypes';
import React from 'react';

interface PromptTemplateCardProps {
    onApply: (prompt: string, promptType: keyof typeof PROMPT_TYPES) => void;
    template: PromptCard;
    prompt: string;
}

export const PromptTemplateCard = React.memo(({ template, prompt, onApply }: PromptTemplateCardProps) => {
    const handleClick = () => {
        onApply(prompt, template.promptType);
    }
    //need to add an onclick functionality to the cards
    return (
     <div>
        <Card onClick={handleClick} className='w-60'>
            <CardHeader>
                <CardTitle>{template.header}</CardTitle>
                <CardDescription>{template.description}</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="w-full aspect-square overflow-hidden rounded">
                    <img src={template.image} alt={template.header} className="w-full h-full object-cover" />
                </div>
            </CardContent>
            <CardFooter>
            </CardFooter>
        </Card>
     </div>   
    )
})