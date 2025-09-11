import React from 'react';

type Props = {
  error: Error;
  resetErrorBoundary: () => void;
};

const ErrorFallback: React.FC<Props> = ({ error, resetErrorBoundary }) => {
  return (
    <div role="alert" className="m-4 rounded-lg border border-red-300 bg-red-50 p-4 text-red-800">
      <div className="font-semibold mb-2">Something went wrong</div>
      <pre className="whitespace-pre-wrap text-sm opacity-80">{error.message}</pre>
      <button
        onClick={resetErrorBoundary}
        className="mt-3 inline-flex items-center rounded-md bg-red-600 px-3 py-1.5 text-white hover:bg-red-700"
      >
        Try again
      </button>
    </div>
  );
};

export default ErrorFallback;


