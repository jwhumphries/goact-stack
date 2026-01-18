import { useState, useEffect } from "react";
import { Button, Card, Spinner } from "@heroui/react";

interface HealthStatus {
  status: string;
}

function App() {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const checkHealth = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/health");
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      setHealth(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to connect");
      setHealth(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkHealth();
  }, []);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <Card.Header>
          <Card.Title>GoAct Stack</Card.Title>
          <Card.Description>
            Full-stack boilerplate with Go + React + Vite + HeroUI
          </Card.Description>
        </Card.Header>
        <Card.Content className="space-y-4">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">Backend Status:</span>
            {loading ? (
              <Spinner size="sm" />
            ) : health ? (
              <span className="text-sm text-green-600 dark:text-green-400">
                {health.status}
              </span>
            ) : (
              <span className="text-sm text-red-600 dark:text-red-400">
                {error || "disconnected"}
              </span>
            )}
          </div>
          <div className="text-sm text-muted">
            <p>Tech Stack:</p>
            <ul className="list-disc list-inside mt-2 space-y-1">
              <li>Go with Echo framework</li>
              <li>React 19 with TypeScript</li>
              <li>Vite for development & bundling</li>
              <li>Tailwind CSS v4</li>
              <li>HeroUI v3 components</li>
            </ul>
          </div>
        </Card.Content>
        <Card.Footer className="flex gap-3">
          <Button onPress={checkHealth} variant="secondary">
            Refresh Status
          </Button>
          <Button
            onPress={() => window.open("https://v3.heroui.com", "_blank")}
            variant="tertiary"
          >
            HeroUI Docs
          </Button>
        </Card.Footer>
      </Card>
    </div>
  );
}

export default App;
