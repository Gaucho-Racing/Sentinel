import React from "react";
import axios from "axios";
import { SENTINEL_API_URL } from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  getAxiosErrorCode,
  getAxiosErrorMessage,
} from "./lib/axios-error-handler";
import { useNavigate } from "react-router-dom";
import { Separator } from "./components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faDiscord } from "@fortawesome/free-brands-svg-icons";

function App() {
  const navigate = useNavigate();

  const [sentinelMsg, setSentinelMsg] = React.useState("");

  const [loginLoading, setLoginLoading] = React.useState(false);
  const [loginEmail, setLoginEmail] = React.useState("");
  const [loginPassword, setLoginPassword] = React.useState("");

  React.useEffect(() => {
    ping();
  }, []);

  const ping = async () => {
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/ping`);
      console.log(response.data);
      setSentinelMsg(response.data.message);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
  };

  const login = async () => {
    setLoginLoading(true);
    try {
      const response = await axios.post(`${SENTINEL_API_URL}/auth/login`, {
        email: loginEmail,
        password: loginPassword,
      });
      if (response.status == 200) {
        localStorage.setItem("id", response.data.data.id);
        localStorage.setItem("token", response.data.data.token);
        navigate("/auth/register");
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    setLoginLoading(false);
  };

  return (
    <>
      <div
        className="flex flex-col items-center justify-between"
        style={{ height: "100vh" }}
      >
        <div className="w-full"></div>
        <div className="p-32">
          <Card className="p-8" style={{ width: 500 }}>
            <div className="items-center">
              <img
                src="logo/mechanic-logo.png"
                alt="Gaucho Racing"
                className="mx-auto h-24"
              />
              <h1 className="mt-6 text-2xl font-semibold tracking-tight">
                Sentinel Sign On
              </h1>
              <Input
                id="email"
                className="mt-4"
                placeholder="Email"
                type="email"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                disabled={loginLoading}
                onChange={(e) => {
                  setLoginEmail(e.target.value);
                }}
              />
              <Input
                className="mt-2"
                id="password"
                placeholder="Password"
                type="password"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                disabled={loginLoading}
                onChange={(e) => {
                  setLoginPassword(e.target.value);
                }}
              />
              <Button
                disabled={loginLoading}
                className="mt-4 w-full"
                onClick={login}
              >
                {loginLoading && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Sign In with Email
              </Button>
              <div className="flex items-center justify-center pt-4 text-xl font-semibold text-gray-500">
                <Separator className="m-4 w-32" />
                <p>OR</p>
                <Separator className="m-4 w-32" />
              </div>
              <button
                className="bg-discord-blurple hover:bg-discord-blurple/90 mt-4 w-full rounded-md p-2 font-medium text-white transition-colors"
                onClick={() => {
                  window.location.href = `https://discord.com/oauth2/authorize?client_id=1204930904913481840&response_type=code&redirect_uri=http%3A%2F%2Flocalhost%3A5173%2Fauth%2Flogin%2Fdiscord&scope=identify+email`;
                }}
              >
                <span className="flex items-center justify-center">
                  <FontAwesomeIcon icon={faDiscord} className="me-2" />
                  Sign In with Discord
                </span>
              </button>
            </div>
          </Card>
        </div>
        <div className="flex w-full justify-end p-4 text-gray-500">
          <p>{sentinelMsg}</p>
        </div>
      </div>
    </>
  );
}

export default App;
