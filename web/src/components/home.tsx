import { Button } from "@/components/ui/button"
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card"
import React from "react";

export default function Home() {
  return (
    <div className="container mx-auto px-4 min-h-screen flex items-center">
      <Card className="max-w-2xl mx-auto border-none">
        <CardHeader>
          <CardTitle className="text-4xl font-semibold text-center">Welcome to Asteroid</CardTitle>
          <CardDescription className="text-xl text-center mt-2">
            Supervision and evaluation for agentic systems
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <p className="mb-6">
            Asteroid is a platform for supervision of agentic systems. We want to make it easy to safely deploy and oversee clusters of millions of agents, that do useful work on the internet. Our mission is to make agentic systems reliable, beneficial and simple.
          </p>
          <div className="flex justify-center gap-4">

            <Button asChild className="mb-4">
              <a href="https://docs.asteroid.ai">
                Documentation
              </a>
            </Button>
            <Button asChild className="mb-4">
              <a href="https://github.com/asteroidai/asteroid">
                GitHub
              </a>
            </Button>
          </div>
        </CardContent>
        <CardFooter className="flex justify-center text-center">
          <p className="text-sm text-muted-foreground">
            We're excited to have you here. If you need anything, please contact us at{' '}
            <a href="mailto:founders@asteroid.ai" className="text-primary hover:underline">
              founders@asteroid.ai
            </a>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}
