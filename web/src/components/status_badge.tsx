import React from "react";
import { Status } from "@/types";
import { Badge } from "@/components/ui/badge";

export function StatusBadge({ status }: { status: Status }) {
  return <Badge>{status}</Badge>;
}
