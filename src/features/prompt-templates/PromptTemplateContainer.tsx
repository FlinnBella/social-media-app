/*
Imports
*/
import { PromptTemplateCard } from './PromptTemplateCard';
import { PromptCard } from './prompttypes';
import { PROMPT_TYPES } from './prompttypes';


interface PromptTemplateContainerProps {
    templates: PromptCard[];
    prompt: string; 
    onApply: (promptType: keyof typeof PROMPT_TYPES) => void;
}

/*
Component scoped functions and state
ReactNode?
*/
export const PromptTemplateContainer = ({templates, prompt, onApply}: PromptTemplateContainerProps)  => {

    //look at how this works in the 
    //timeline cards component. 
    const MapTemplateCards = (
        templates: PromptCard[]
    ) => {

        const transformedTempaltes = [...templates];
        return transformedTempaltes.map((template, index) => {
            
                    return(
                        <div className="flex" key={index}>
                            <PromptTemplateCard key={index} template={template} prompt={prompt} onApply={onApply} />
                        </div>
                    )
                })
        }

    return(
            <div className='flex flex-row gap-4'>
                {MapTemplateCards(templates)}
            </div>
    )
};










/*
JSX
*/