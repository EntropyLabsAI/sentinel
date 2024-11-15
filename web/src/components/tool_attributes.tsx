import React from "react";
import { ToolAttributes as ToolAttributesType } from "@/types";

type ToolAttributesProps = {
  attributes: ToolAttributesType;
  ignoredAttributes: string[];
};

export function ToolAttributes({ attributes, ignoredAttributes }: ToolAttributesProps) {
  if (Object.keys(attributes).length === 0) {
    return null;
  }

  return <pre className="text-xs bg-muted p-2 rounded overflow-scroll">{JSON.stringify(attributes, null, 2)}</pre>;
}
