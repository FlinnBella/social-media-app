import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import '@fontsource/ibm-plex-sans/400.css';
import '@fontsource/ibm-plex-sans/600.css';
import '@fontsource/ibm-plex-sans/700.css';
import App from './App.tsx';
import './index.css';
import { Toaster } from 'sonner';
import { SubmissionProvider } from '@/context/SubmissionContext';
import { ErrorBoundary } from 'react-error-boundary';
import ErrorFallback from '@/components/Error/ErrorFallback.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, info) => {
        // eslint-disable-next-line no-console
        console.error('App crashed:', error, info);
      }}
      onReset={() => {
        // no-op; could reset global state here if needed
      }}
    >
      <SubmissionProvider>
        <App />
      </SubmissionProvider>
    </ErrorBoundary>
    <Toaster />
  </StrictMode>
);
