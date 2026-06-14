"use client";
import { useRef, useState } from "react";
import { ArrowUpRightFromSquare, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

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
          className="border px-1 border-input rounded-sm w-full outline-none bg-muted text-muted-foreground"
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
          <div className="absolute inset-0 flex items-center justify-center bg-muted">
            <div className="text-center p-4">
              <p className="text-destructive">{error}</p>
              <Button onClick={handleRefresh} size="sm" className="mt-2">
                Retry
              </Button>
            </div>
          </div>
        )}

        {isLoading && !error && (
          <div className="absolute inset-0 flex items-center justify-center bg-muted">
            <div className="animate-pulse text-muted-foreground">Loading...</div>
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
