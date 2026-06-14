"use client";
import { Button } from "@/components/ui/button";
import { LogOut } from "lucide-react";
import { ModeToggle } from "../ModeToggle";
import Link from "next/link";

const Navbar = ({ name }: { name: string }) => {
  return (
    <div className="border-b p-2 px-4 flex justify-between items-center bg-background">
      <span className="font-semibold">{name}</span>
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
