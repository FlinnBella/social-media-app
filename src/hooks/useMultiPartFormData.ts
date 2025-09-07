import { toast } from "react-toastify";
enum Action {

    SendImageTimeline = 'SendImageTimeline',
    imageTimeline = 'imageTimeline',
    finalVideo = 'finalVideo'
}

// take in the actual formdata, what action is occuring
//then perform that keeping the server url in mind
export const useMultiPartFormData = (FormData: any, action: Action, serverUrl: string) => {

    //use this to handle all form data operations
    switch (action) {
        case Action.SendImageTimeline:
            return SendImageTimeline(FormData, serverUrl);
        case Action.imageTimeline:
            return imageTimeline(FormData, serverUrl);
            case Action.finalVideo:
                return finalVideo(FormData, serverUrl);
        default: 
        console.error('Invalid form data type');
        return toast.error('Invalid form data type');
    }


    //these will handle sending and receiving the form-data; 
    //put all the network logic in some function calls
    function SendImageTimeline(FormData: any, serverUrl: string) {
        return FormData;
    }
    function imageTimeline(FormData: any, serverUrl: string) {
        return FormData;
    }
    function finalVideo(FormData: any, serverUrl: string) {
        return FormData;
    }

}
