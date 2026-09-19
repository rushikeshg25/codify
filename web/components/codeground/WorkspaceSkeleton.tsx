import { Loader2 } from "lucide-react";

export default function WorkspaceSkeleton() {
  return (
    <div className="h-screen flex items-center justify-center gap-2 text-sm text-muted-foreground">
      <Loader2 className="w-4 h-4 animate-spin" />
      Loading codeground…
    </div>
  );
}
