"use client";
import { useRef, useState } from "react";
import { ArrowUpRightFromSquare, RefreshCw } from "lucide-react";

const Webview = ({ codegroundId }: { codegroundId: string }) => {
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleOpenTab = () => {
    if (iframeRef.current) {
      window.open(iframeRef.current.src, "_blank");
    }
  };

  const handleRefresh = () => {
    if (iframeRef.current) {
      setIsLoading(true);
      setError(null);
      iframeRef.current.src = `http://app-${codegroundId}.codify.localhost`;
    }
  };

  const handleIframeLoad = () => {
    setIsLoading(false);
  };

  const handleIframeError = () => {
    setError(
      "Failed to load content. Please ensure the development server is running.",
    );
    setIsLoading(false);
  };

  return (
    <div className="flex flex-col h-screen">
      <div className="flex items-center justify-between p-2 space-x-2 border-b">
        <RefreshCw
          className={`cursor-pointer h-5 w-5 ${isLoading ? "animate-spin" : ""}`}
          onClick={handleRefresh}
          role="button"
        />
        <input
          className="border px-1 border-gray-300 dark:border-gray-700 rounded-sm w-full outline-none bg-gray-50 dark:bg-gray-800"
          value={`app-${codegroundId}.codify.localhost`}
          type="text"
          disabled
        />
        <ArrowUpRightFromSquare
          className="cursor-pointer h-4 w-4"
          onClick={handleOpenTab}
          role="button"
        />
      </div>

      <div className="flex-1 relative">
        {error && (
          <div className="absolute inset-0 flex items-center justify-center bg-gray-50 dark:bg-gray-800">
            <div className="text-red-500 text-center p-4">
              <p>{error}</p>
              <button
                onClick={handleRefresh}
                className="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                Retry
              </button>
            </div>
          </div>
        )}

        {isLoading && !error && (
          <div className="absolute inset-0 flex items-center justify-center bg-gray-50 dark:bg-gray-800">
            <div className="animate-pulse text-gray-500">Loading...</div>
          </div>
        )}

        <iframe
          sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          src={`http://app-${codegroundId}.codify.localhost`}
          className="w-full h-full"
          ref={iframeRef}
          onLoad={handleIframeLoad}
          onError={handleIframeError}
        />
      </div>
    </div>
  );
};

export default Webview;
