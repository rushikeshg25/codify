import React from "react";
import { Button } from "@/components/ui/button";

import { Files, Play, Settings } from "lucide-react";
import { ModeToggle } from "../ModeToggle";

const Navbar = () => {
  return (
    <div className="border-b p-2 flex justify-between items-center bg-background">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon">
          <Files className="w-4 h-4" />
        </Button>
        <span className="font-semibold">React Todo App</span>
      </div>
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" className="gap-2">
          <Play className="w-4 h-4" />
          Run
        </Button>
        <Button variant="ghost" size="icon">
          <Settings className="w-4 h-4" />
        </Button>
        <ModeToggle />
      </div>
    </div>
  );
};

export default Navbar;
