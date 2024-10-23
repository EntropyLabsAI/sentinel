import { Button } from "@/components/ui/button"
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card"
import { BookAIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";

export default function Home() {
  return (
    <div className="container mx-auto my-auto px-4 py-16">
      <Card className="max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle className="text-4xl font-semibold text-center">Welcome to Sentinel</CardTitle>
          <CardDescription className="text-xl text-center mt-2">
            Supervision and evaluation for agentic systems
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <p className="mb-6">
            Sentinel is an experimental platform for supervision of agentic systems. We want to make it easy to safely deploy and oversee clusters of millions of agents, that do useful work on the internet. Our mission is to make agentic systems reliable, beneficial and simple.
          </p>
          <div className="flex justify-center gap-4">

            <Button asChild className="mb-4">
              <a href="https://docs.entropy-labs.ai">
                Documentation
              </a>
            </Button>
            <Button asChild className="mb-4">
              <a href="https://github.com/EntropyLabsAI/sentinel">
                GitHub
              </a>
            </Button>
          </div>
        </CardContent>
        <CardFooter className="flex justify-center text-center">
          <p className="text-sm text-muted-foreground">
            We're excited to have you here. If you need anything, please contact us at{' '}
            <a href="mailto:devs@entropy-labs.ai" className="text-primary hover:underline">
              devs@entropy-labs.ai
            </a>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}
