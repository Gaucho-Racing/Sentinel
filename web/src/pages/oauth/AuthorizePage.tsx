import React from "react";
import axios from "axios";
import { SENTINEL_API_URL } from "@/consts/config";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useSearchParams } from "react-router-dom";
import { checkCredentials } from "@/lib/auth";
import { ClientApplication, initClientApplication } from "@/models/application";
import { OutlineButton } from "@/components/ui/outline-button";
import { Button } from "@/components/ui/button";
import { notify } from "@/lib/notify";

function AuthorizePage() {
  const navigate = useNavigate();
  const [queryParameters] = useSearchParams();

  const [sentinelMsg, setSentinelMsg] = React.useState("");
  const [loginLoading, setLoginLoading] = React.useState(true);
  const [errorMsg, setErrorMsg] = React.useState("");
  const [promptRequired, setPromptRequired] = React.useState(false);

  const [application, setApplication] = React.useState<ClientApplication>(
    initClientApplication,
  );

  React.useEffect(() => {
    checkAuth().then(() => {
      ping();
      validate();
    });
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
    const currentRoute = window.location.pathname + window.location.search;
    const status = await checkCredentials();
    if (status != 0) {
      navigate(`/auth/login?route=${encodeURIComponent(currentRoute)}`);
    }
  };

  const validate = async () => {
    setLoginLoading(true);
    const url = window.location.href;
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/oauth/authorize${url.split("oauth/authorize")[1]}`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      if (response.status == 200) {
        setErrorMsg("");
        getApplication(response.data.client_id);
        if (response.data.prompt == "consent") {
          setLoginLoading(false);
          setPromptRequired(true);
        } else if (response.data.prompt == "none") {
          setPromptRequired(false);
          authorize();
        }
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setLoginLoading(false);
      setErrorMsg(error.response.data.message);
    }
  };

  const authorize = async () => {
    setLoginLoading(true);
    const url = window.location.href;
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/oauth/authorize${url.split("oauth/authorize")[1]}`,
        {},
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      if (response.status == 200) {
        handleRedirect(response.data.code);
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setLoginLoading(false);
      setErrorMsg(error.response.data.message);
    }
  };

  const getApplication = async (clientId: string) => {
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/applications/${clientId}`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setApplication(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
  };

  const handleRedirect = (code: string) => {
    const state = queryParameters.get("state");
    const redirectUri = queryParameters.get("redirect_uri");
    window.location.href = redirectUri + `?code=${code}&state=${state}`;
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
            Sentinel OAuth Error
          </h1>
          <p className="mt-4">{errorMsg}</p>
        </div>
      </Card>
    );
  };

  const PromptCard = () => {
    return (
      <Card className="p-4 md:w-[500px] md:p-8">
        <div className="items-center">
          <img
            src="/logo/mechanic-logo.png"
            alt="Gaucho Racing"
            className="mx-auto h-20 md:h-24"
          />
          <h1 className="mt-6 text-2xl font-semibold tracking-tight">
            Login to {application.name}
          </h1>
          <p className="mt-4">
            {application.name} would like to access your Sentinel account.
          </p>
          <p className="mt-4">Requested Scopes:</p>
          <div className="mt-2 flex flex-wrap">
            {queryParameters
              .get("scope")
              ?.split(" ")
              .map((scope) => (
                <div key={scope} className="mx-1 mb-2">
                  <Card className="rounded-sm px-1 text-gray-400">
                    <code className="">{scope}</code>
                  </Card>
                </div>
              ))}
          </div>
          <div className="mt-4 flex w-full items-center justify-end">
            <Button
              className="mr-2"
              variant={"ghost"}
              onClick={() => {
                navigate("/");
              }}
            >
              Cancel
            </Button>
            <OutlineButton
              className=""
              onClick={() => {
                authorize();
              }}
            >
              Authorize
            </OutlineButton>
          </div>
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
          ) : promptRequired ? (
            <PromptCard />
          ) : errorMsg != "" ? (
            <InvalidCodeCard />
          ) : (
            <></>
          )}
        </div>
        <div className="flex w-full justify-end p-4 text-gray-500">
          <p>{sentinelMsg}</p>
        </div>
      </div>
    </>
  );
}

export default AuthorizePage;
