import * as React from "react";
import { Button } from "./button";

export interface FileIconProps {
  file: File;
  setFiles: React.Dispatch<React.SetStateAction<File[]>>;
}

const FileIcon = React.forwardRef<HTMLButtonElement, FileIconProps>(
  ({ file, setFiles, ...props }, ref) => {
    return (
      <div className="flex flex-row gap-2 bg-slate-500 text-sm rounded-md p-2 text-slate-200 w-96">
        <div className="w-80 overflow-hidden">
          <span>{file.name}</span>
          <span>{file.type}</span>
          <span>{file.size}</span>
        </div>
        <div className="20">
          <Button
            {...props}
            ref={ref}
            className="bg-slate-700 text-slate-200"
            onClick={() => {
              setFiles((prev) => prev.filter((f) => f !== file));
            }}
          >
            <span className="">X</span>
          </Button>
        </div>
      </div>
    );
  }
);

export { FileIcon };
