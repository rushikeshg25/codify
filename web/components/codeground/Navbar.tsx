"use client";
import { Button } from "@/components/ui/button";
import { Files, Play, Settings } from "lucide-react";
import { ModeToggle } from "../ModeToggle";
import { LogOut } from "lucide-react";
import Link from "next/link";

const Navbar = () => {
  return (
    <div className="border-b p-2 px-4 flex justify-between items-center bg-background">
      <div className="flex items-center gap-2">
        <span className="font-semibold">React Todo App</span>
      </div>
      <Button size="sm" variant="outline" className="gap-2">
        <Play className="w-4 h-4" />
        Run
      </Button>
      <div className="flex items-center gap-2">
        <ModeToggle />

        <Button size="sm" variant="outline" className="gap-2" asChild>
          <Link href="/codeground">
            <LogOut className="w-4 h-4" />
            Exit Codeground
          </Link>
        </Button>
      </div>
    </div>
  );
};

export default Navbar;
