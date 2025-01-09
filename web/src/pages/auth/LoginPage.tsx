import React from "react";
import axios from "axios";
import {
  DISCORD_CLIENT_ID,
  DISCORD_OAUTH_BASE_URL,
  SENTINEL_API_URL,
} from "@/consts/config";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faDiscord } from "@fortawesome/free-brands-svg-icons";
import { checkCredentials, saveAccessToken, saveRefreshToken } from "@/lib/auth";
import { OutlineButton } from "@/components/ui/outline-button";
import { notify } from "@/lib/notify";

function LoginPage() {
  const navigate = useNavigate();
  const [queryParameters] = useSearchParams();

  const [sentinelMsg, setSentinelMsg] = React.useState("");

  const [loginLoading, setLoginLoading] = React.useState(false);
  const [loginEmail, setLoginEmail] = React.useState("");
  const [loginPassword, setLoginPassword] = React.useState("");

  React.useEffect(() => {
    ping();
    checkAuth();
  }, []);

  const ping = async () => {
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/ping`);
      console.log(response.data);
      setSentinelMsg(response.data.message);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
  };

  const checkAuth = async () => {
    const status = await checkCredentials();
    console.log(status);
    if (status == 0) {
      handleRedirect();
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
        saveAccessToken(response.data.access_token);
        saveRefreshToken(response.data.refresh_token);
        checkAuth();
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    setLoginLoading(false);
  };

  const handleRedirect = () => {
    const route = queryParameters.get("route");
    if (route) {
      navigate(route);
    } else {
      navigate("/");
    }
  };

  return (
    <>
      <div className="flex h-screen flex-col items-center justify-between">
        <div className="w-full"></div>
        <div className="w-full items-center justify-center p-4 md:flex md:p-32">
          <Card className="p-4 md:w-[500px] md:p-8">
            <div className="items-center">
              <img
                src="/logo/mechanic-logo.png"
                alt="Gaucho Racing"
                className="mx-auto h-20 md:h-24"
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
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    login();
                  }
                }}
              />
              <OutlineButton
                disabled={loginLoading}
                className="mt-4 w-full"
                onClick={login}
              >
                {loginLoading && <Loader2 className="mr-2 animate-spin" />}
                Sign In with Email
              </OutlineButton>
              <div className="flex items-center justify-center pt-4 text-xl font-semibold text-gray-500">
                <Separator className="m-4 w-8 md:w-32" />
                <p>OR</p>
                <Separator className="m-4 w-8 md:w-32" />
              </div>
              <button
                className="mt-4 w-full rounded-md bg-discord-blurple p-2 font-medium text-white transition-colors hover:bg-discord-blurple/90"
                onClick={() => {
                  const redirect_url =
                    window.location.origin + "/auth/login/discord";
                  const scope = "identify+email";
                  let oauthUrl = `${DISCORD_OAUTH_BASE_URL}?client_id=${DISCORD_CLIENT_ID}&response_type=code&redirect_uri=${encodeURIComponent(redirect_url)}&scope=${scope}&prompt=none`;
                  const route = queryParameters.get("route");
                  if (route) {
                    oauthUrl += `&state=${encodeURIComponent(route)}`;
                  }
                  window.location.href = oauthUrl;
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

export default LoginPage;
