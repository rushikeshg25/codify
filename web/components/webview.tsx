import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";

const Webview = () => {
  return (
    <ScrollArea className="p-4 font-mono text-sm flex items-center justify-center">
      <div>Webview here</div>
      <ScrollBar orientation="horizontal" />
    </ScrollArea>
  );
};

export default Webview;
