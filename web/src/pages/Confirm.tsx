import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/auth/useAuth";
import { Loader2, CheckCircle, XCircle } from "lucide-react";
import { PATHS } from "@/router/paths";

export default function Confirm() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { activate } = useAuth();

  const [status, setStatus] = useState<"loading" | "success" | "error">(
    "loading",
  );

  useEffect(() => {
    if (!token) {
      setStatus("error");
      return;
    }

    activate(token)
      .then(() => {
        setStatus("success");
        setTimeout(() => navigate(PATHS.HOME), 3000);
      })
      .catch(() => {
        setStatus("error");
      });
  }, [token, activate, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 text-foreground">
      <div className="flex flex-col items-center space-y-4 text-center rounded-3xl border border-border p-10 shadow-2xl bg-card">
        {status === "loading" && (
          <>
            <Loader2 className="h-12 w-12 animate-spin text-primary" />
            <h2 className="text-2xl font-semibold">Activating Account...</h2>
          </>
        )}
        {status === "success" && (
          <>
            <CheckCircle className="h-12 w-12 text-green-500" />
            <h2 className="text-2xl font-semibold">
              Account Activated Successfully!
            </h2>
            <p className="text-sm text-muted-foreground">
              Redirecting to home...
            </p>
          </>
        )}
        {status === "error" && (
          <>
            <XCircle className="h-12 w-12 text-destructive" />
            <h2 className="text-2xl font-semibold">Activation Failed</h2>
            <p className="text-sm text-muted-foreground">
              The link is invalid or has expired.
            </p>
          </>
        )}
      </div>
    </div>
  );
}
