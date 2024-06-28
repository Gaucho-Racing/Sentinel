import React from "react";
import axios from "axios";
import { SENTINEL_API_URL } from "@/consts/config";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  getAxiosErrorCode,
  getAxiosErrorMessage,
} from "@/lib/axios-error-handler";
import { useNavigate } from "react-router-dom";

function LoginDiscordPage() {
  const navigate = useNavigate();

  const [sentinelMsg, setSentinelMsg] = React.useState("");

  const [loginLoading, setLoginLoading] = React.useState(true);

  //   React.useEffect(() => {
  //     ping();
  //   }, []);

  React.useEffect(() => {
    console.log("Calling login function");
    login();
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
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get("code");
    if (!code) {
      navigate("/");
      return;
    }
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/auth/login/discord?code=${code}`,
      );
      if (response.status == 200) {
        localStorage.setItem("id", response.data.id);
        localStorage.setItem("token", response.data.token);
        navigate("/");
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
  };

  return (
    <>
      <div
        className="flex flex-col items-center justify-between"
        style={{ height: "100vh" }}
      >
        <div className="w-full"></div>
        <div className="p-32">
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
        <div className="flex w-full justify-end p-4 text-gray-500">
          <p>{sentinelMsg}</p>
        </div>
      </div>
    </>
  );
}

export default LoginDiscordPage;
