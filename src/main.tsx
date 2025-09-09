import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import '@fontsource/ibm-plex-sans/400.css';
import '@fontsource/ibm-plex-sans/600.css';
import '@fontsource/ibm-plex-sans/700.css';
import App from './App.tsx';
import './index.css';
import { Toaster } from 'sonner';
import { SubmissionProvider } from '@/context/SubmissionContext';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <SubmissionProvider>
      <App />
    </SubmissionProvider>
    <Toaster />
  </StrictMode>
);
