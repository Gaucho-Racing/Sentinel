import React from "react";
import axios from "axios";
import {
  DISCORD_CLIENT_ID,
  DISCORD_OAUTH_BASE_URL,
  SENTINEL_API_URL,
  currentUser,
} from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { getAxiosErrorMessage } from "./lib/axios-error-handler";
import { useNavigate } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faDiscord } from "@fortawesome/free-brands-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";

function App() {
  const navigate = useNavigate();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  React.useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    setAuthCheckLoading(true);
    const currentRoute = window.location.pathname + window.location.search;
    const status = await checkCredentials();
    if (status != 0) {
      navigate(`/auth/login?route=${encodeURIComponent(currentRoute)}`);
    } else {
      setAuthCheckLoading(false);
    }
  };

  const AuthLoading = () => {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Card className="border-none p-8" style={{ width: 500 }}>
          <div className="flex flex-col items-center justify-center">
            <img
              src="/logo/mechanic-logo.png"
              alt="Gaucho Racing"
              className="mx-auto h-24"
            />
            <Loader2 className="mt-8 h-16 w-16 animate-spin" />
          </div>
        </Card>
      </div>
    );
  };

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col items-center justify-between">
          <div className="w-full"></div>
          <div className="p-32">{currentUser.first_name}</div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default App;
