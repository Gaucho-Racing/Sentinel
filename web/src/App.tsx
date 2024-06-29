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
import { faPerson, faUser } from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import Header from "@/components/Header";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";

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

  const ProfileField = (props: { label: string; value: string }) => {
    return (
      <div className="mx-2 mt-2 flex">
        <div className="mr-2 font-semibold">{props.label}:</div>
        <div className="text-gray-400">
          {props.value != "" ? props.value : "Not set"}
        </div>
      </div>
    );
  };

  const ProfileCard = () => {
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faUser} className="h-5 w-5" />
          <h3 className="ml-4">Profile</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-start">
          <Avatar className="mr-4">
            <AvatarImage src={currentUser.avatar_url} />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <p>
              {currentUser.first_name} {currentUser.last_name}
            </p>
            <p className="text-gray-400">{currentUser.email}</p>
          </div>
        </div>
        <ProfileField label="ID" value={currentUser.id} />
        <ProfileField label="Email" value={currentUser.email} />
        <ProfileField label="Phone Number" value={currentUser.phone_number} />
        <ProfileField
          label="Graduate Level"
          value={currentUser.graduate_level}
        />
        <ProfileField
          label="Graduate Year"
          value={currentUser.graduation_year.toString()}
        />
        <ProfileField label="Major" value={currentUser.major} />
        <ProfileField label="Shirt Size" value={currentUser.shirt_size} />
        <ProfileField label="Jacket Size" value={currentUser.jacket_size} />
        <ProfileField
          label="SAE Member Number"
          value={currentUser.sae_registration_number}
        />
        <ProfileField
          label="Subteams"
          value={currentUser.subteams.map((subteam) => subteam.name).join(", ")}
        />
        <div className="mx-2 mt-2 flex">
          <div className="mr-2 font-semibold">Roles:</div>
          <div className="flex flex-wrap">
            {currentUser.roles.map((role) => (
              <div key={role} className="mx-1 mb-2">
                <Card className="rounded-sm px-1 text-gray-400">
                  <code className="">{role}</code>
                </Card>
              </div>
            ))}
          </div>
        </div>
        <ProfileField
          label="Updated At"
          value={new Date(currentUser.updated_at).toLocaleString()}
        />
        <ProfileField
          label="Created At"
          value={new Date(currentUser.created_at).toLocaleString()}
        />
      </Card>
    );
  };

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col justify-between">
          <div className="p-4 lg:p-32 lg:pt-16">
            <h1>Hello {currentUser.first_name}</h1>
            <div className="flex flex-wrap">
              <ProfileCard />
              <ProfileCard />
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default App;
