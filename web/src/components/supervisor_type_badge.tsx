import React from "react";
import { Badge } from "./ui/badge";
import { SupervisorType } from "@/types";

export const SupervisorTypeBadge: React.FC<{ type: SupervisorType }> = ({ type }) => {
  const label = type === SupervisorType.client_supervisor ? "Client-side Supervision" : "Human Supervision";
  const color = type === SupervisorType.client_supervisor ? "blue" : "green";
  return <Badge className={`text-white bg-${color}-500`} variant="outline">{label}</Badge>;
};
