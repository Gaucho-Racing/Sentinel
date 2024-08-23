import React from "react";
import axios from "axios";
import {
  DISCORD_CLIENT_ID,
  DISCORD_OAUTH_BASE_URL,
  DISCORD_SERVER_INVITE_URL,
  SENTINEL_API_URL,
} from "@/consts/config";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useSearchParams } from "react-router-dom";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faDiscord } from "@fortawesome/free-brands-svg-icons";
import { checkCredentials, saveAccessToken } from "@/lib/auth";
import { notify } from "@/lib/notify";

function LoginDiscordPage() {
  const navigate = useNavigate();
  const [queryParameters] = useSearchParams();

  const [sentinelMsg, setSentinelMsg] = React.useState("");
  const [loginLoading, setLoginLoading] = React.useState(true);
  const [accountExists, setAccountExists] = React.useState(true);

  React.useEffect(() => {
    ping();
    login();
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
    if (status == 0) {
      handleRedirect();
    }
  };

  const login = async () => {
    const code = queryParameters.get("code");
    if (!code) {
      navigate("/");
      return;
    }
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/auth/login/discord?code=${code}`,
      );
      if (response.status == 200) {
        saveAccessToken(response.data.access_token);
        checkAuth();
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setLoginLoading(false);
      if (getAxiosErrorMessage(error).includes("No account with this")) {
        setAccountExists(false);
      }
    }
  };

  const handleRedirect = () => {
    const route = queryParameters.get("state");
    if (route) {
      navigate(route);
    } else {
      navigate("/");
    }
  };

  const LoadingCard = () => {
    return (
      <Card className="border-none p-4 md:w-[500px] md:p-8">
        <div className="flex flex-col items-center justify-center">
          <img
            src="/logo/mechanic-logo.png"
            alt="Gaucho Racing"
            className="mx-auto h-20 md:h-24"
          />
          <Loader2 className="mt-8 h-16 w-16 animate-spin" />
        </div>
      </Card>
    );
  };

  const InvalidCodeCard = () => {
    return (
      <Card className="p-4 md:w-[500px] md:p-8">
        <div className="items-center">
          <img
            src="/logo/mechanic-logo.png"
            alt="Gaucho Racing"
            className="mx-auto h-20 md:h-24"
          />
          <h1 className="mt-6 text-2xl font-semibold tracking-tight">
            Discord OAuth Error
          </h1>
          <p className="mt-4">Invalid or expired code. Please try again.</p>
          <button
            className="mt-4 w-full rounded-md bg-discord-blurple p-2 font-medium text-white transition-colors hover:bg-discord-blurple/90"
            onClick={() => {
              const redirect_url =
                window.location.origin + "/auth/login/discord";
              const scope = "identify+email";
              let oauthUrl = `${DISCORD_OAUTH_BASE_URL}?client_id=${DISCORD_CLIENT_ID}&response_type=code&redirect_uri=${encodeURIComponent(redirect_url)}&scope=${scope}`;
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
    );
  };

  const NoAccountCard = () => {
    return (
      <Card className="p-4 md:w-[500px] md:p-8">
        <div className="items-center">
          <img
            src="/logo/mechanic-logo.png"
            alt="Gaucho Racing"
            className="mx-auto h-20 md:h-24"
          />
          <h1 className="mt-6 text-2xl font-semibold tracking-tight">
            No Account Found
          </h1>
          <p className="mt-4">
            No Sentinel account found. Make sure that you have joined the Gaucho
            Racing Discord server and verified your account.
          </p>
          <p className="mt-4">
            You can verify your account using the <code>!verify</code> command
            in the <strong>#verification</strong> channel.
            <br />
            <br />
            Example: <code>{`!verify <first name> <last name> <email>`}</code>
          </p>
          <button
            className="mt-4 w-full rounded-md bg-discord-blurple p-2 font-medium text-white transition-colors hover:bg-discord-blurple/90"
            onClick={() => {
              window.location.href = DISCORD_SERVER_INVITE_URL;
            }}
          >
            <span className="flex items-center justify-center">
              <FontAwesomeIcon icon={faDiscord} className="me-2" />
              Join the Discord
            </span>
          </button>
        </div>
      </Card>
    );
  };

  return (
    <>
      <div className="flex h-screen flex-col items-center justify-between">
        <div className="w-full"></div>
        <div className="w-full items-center justify-center p-4 md:flex md:p-32">
          {loginLoading ? (
            <LoadingCard />
          ) : accountExists ? (
            <InvalidCodeCard />
          ) : (
            <NoAccountCard />
          )}
        </div>
        <div className="flex w-full justify-end p-4 text-gray-500">
          <p>{sentinelMsg}</p>
        </div>
      </div>
    </>
  );
}

export default LoginDiscordPage;
