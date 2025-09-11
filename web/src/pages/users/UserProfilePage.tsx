import React from "react";
import axios from "axios";
import { useNavigate, useParams } from "react-router-dom";
import { SENTINEL_API_URL } from "@/consts/config";
import { checkCredentials } from "@/lib/auth";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { notify } from "@/lib/notify";
import { useUser } from "@/lib/store";
import { User, initUser, isInnerCircle } from "@/models/user";
import Footer from "@/components/Footer";
import { AuthLoading } from "@/components/AuthLoading";
import { Button } from "@/components/ui/button";
import { OutlineButton } from "@/components/ui/outline-button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Loader2 } from "lucide-react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowLeft, faChartLine, faClockRotateLeft, faUser, faMessage, faFaceSmile } from "@fortawesome/free-solid-svg-icons";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

type UserLogin = {
  id: string;
  user_id: string;
  destination: string;
  scope: string;
  ip_address: string;
  login_type: string;
  created_at: string;
};

type ActivityCount = {
  date: string;
  action: string;
  count: number;
};

function UserProfilePage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const currentUser = useUser();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);
  const [userLoading, setUserLoading] = React.useState(false);
  const [user, setUser] = React.useState<User>(initUser);
  const [loginsLoading, setLoginsLoading] = React.useState(false);
  const [logins, setLogins] = React.useState<UserLogin[]>([]);
  const [activitiesLoading, setActivitiesLoading] = React.useState(false);
  const [activityStats, setActivityStats] = React.useState<ActivityCount[]>([]);

  React.useEffect(() => {
    checkAuth().then(() => {
      getUser();
      getLogins();
      getActivityStats();
    });
  }, [id]);

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

  const getActivityStats = async () => {
    setActivitiesLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/${id}/activity-stats`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) setActivityStats(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setActivityStats([]);
    }
    setActivitiesLoading(false);
  };

  const getUser = async () => {
    setUserLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/${id}` ,{
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) {
        setUser(response.data);
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setUser(initUser);
    }
    setUserLoading(false);
  };

  const getLogins = async () => {
    setLoginsLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/${id}/logins`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) {
        setLogins(response.data);
      }
    } catch (error: any) {
      // If not authorized, just show none
      if (!getAxiosErrorMessage(error).toLowerCase().includes("unauthorized")) {
        notify.error(getAxiosErrorMessage(error));
      }
      setLogins([]);
    }
    setLoginsLoading(false);
  };

  const canEdit = React.useMemo(() => {
    return isInnerCircle(currentUser) || currentUser.id == id;
  }, [currentUser, id]);

  const lastLogin = React.useMemo(() => {
    if (logins.length === 0) return undefined;
    return [...logins].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())[0];
  }, [logins]);

  const loginSeries = React.useMemo(() => {
    const buckets: Record<string, number> = {};
    const now = new Date();
    for (let i = 29; i >= 0; i--) {
      const d = new Date(now);
      d.setDate(now.getDate() - i);
      const key = d.toISOString().slice(0, 10);
      buckets[key] = 0;
    }
    logins.forEach((l) => {
      const key = new Date(l.created_at).toISOString().slice(0, 10);
      if (buckets[key] !== undefined) buckets[key] += 1;
    });
    return Object.entries(buckets).map(([date, count]) => ({ date, count }));
  }, [logins]);

  const messageSeries = React.useMemo(() => {
    const map: Record<string, number> = {};
    activityStats.filter((a) => a.action === "message").forEach((a) => (map[a.date] = a.count));
    return Object.entries(map).map(([date, count]) => ({ date, count }));
  }, [activityStats]);

  const reactionSeries = React.useMemo(() => {
    const map: Record<string, number> = {};
    activityStats.filter((a) => a.action === "reaction").forEach((a) => (map[a.date] = a.count));
    return Object.entries(map).map(([date, count]) => ({ date, count }));
  }, [activityStats]);

  const ProfileField = (props: { label: string; value: string }) => {
    return (
      <div className="mx-2 mt-2 flex">
        <div className="mr-2 font-semibold">{props.label}:</div>
        <div className="text-gray-400">{props.value != "" ? props.value : "Not set"}</div>
      </div>
    );
  };

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col justify-between">
          <div className="p-4 lg:p-32 lg:pt-16">
            <div className="mb-4">
              <Button variant={"ghost"} onClick={() => navigate("/users")} className="flex items-center">
                <FontAwesomeIcon icon={faArrowLeft} className="mr-2 h-4 w-4 text-gray-400" />
                Back to users
              </Button>
            </div>
            <h1>
              {user.first_name} {user.last_name}
            </h1>
            <div className="flex flex-wrap">
              <Card className="mr-4 mt-4 w-[500px] p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <FontAwesomeIcon icon={faUser} className="h-5 w-5" />
                    <h3 className="ml-4">Profile</h3>
                  </div>
                  {canEdit && (
                    <OutlineButton onClick={() => navigate(`/users/${user.id}/edit`)}>Edit</OutlineButton>
                  )}
                </div>
                <Separator className="my-2" />
                {userLoading ? (
                  <div className="flex w-full justify-center p-4">
                    <Loader2 className="animate-spin" />
                  </div>
                ) : (
                  <>
                    <div className="flex items-center justify-start">
                      <Avatar className="mr-4">
                        <AvatarImage src={user.avatar_url} />
                        <AvatarFallback>CN</AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <p>
                          {user.first_name} {user.last_name}
                        </p>
                        <p className="text-gray-400">{user.email}</p>
                      </div>
                    </div>
                    <ProfileField label="ID" value={user.id} />
                    <ProfileField label="Phone Number" value={user.phone_number} />
                    <ProfileField label="Gender" value={user.gender} />
                    <ProfileField label="Birthday" value={user.birthday} />
                    <ProfileField label="Graduate Level" value={user.graduate_level} />
                    <ProfileField label="Graduation Year" value={user.graduation_year ? user.graduation_year.toString() : ""} />
                    <ProfileField label="Major" value={user.major} />
                    <ProfileField label="Shirt Size" value={user.shirt_size} />
                    <ProfileField label="Jacket Size" value={user.jacket_size} />
                    <ProfileField label="SAE Member Number" value={user.sae_registration_number} />
                    <ProfileField label="Subteams" value={user.subteams.map((s) => s.name).join(", ")} />
                    <div className="mx-2 mt-2 flex">
                      <div className="mr-2 font-semibold">Roles:</div>
                      <div className="flex flex-wrap">
                        {user.roles.map((role) => (
                          <div key={role} className="mx-1 mb-2">
                            <Card className="rounded-sm px-1 text-gray-400">
                              <code className="">{role}</code>
                            </Card>
                          </div>
                        ))}
                      </div>
                    </div>
                    <ProfileField label="Updated At" value={user.updated_at ? new Date(user.updated_at).toLocaleString() : ""} />
                    <ProfileField label="Created At" value={user.created_at ? new Date(user.created_at).toLocaleString() : ""} />
                  </>
                )}
              </Card>

              <div className="mr-4 mt-4 w-[700px]">
                <Card className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <FontAwesomeIcon icon={faClockRotateLeft} className="h-5 w-5" />
                      <h3 className="ml-4">Login History</h3>
                    </div>
                  </div>
                  <Separator className="my-2" />
                  {loginsLoading ? (
                    <div className="flex w-full justify-center p-4">
                      <Loader2 className="animate-spin" />
                    </div>
                  ) : (
                    <div>
                      <div className="mb-4 flex gap-8 text-sm text-gray-300">
                        <div>
                          <span className="font-semibold">Total Logins:</span> {logins.length}
                        </div>
                        <div>
                          <span className="font-semibold">Last Login:</span> {lastLogin ? new Date(lastLogin.created_at).toLocaleString() : "—"}
                        </div>
                      </div>
                      <div className="h-64 w-full">
                        <ResponsiveContainer width="100%" height="100%">
                          <LineChart data={loginSeries} margin={{ left: 8, right: 16, top: 8, bottom: 8 }}>
                            <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                            <XAxis dataKey="date" tick={{ fill: "#9ca3af", fontSize: 12 }} hide={false} />
                            <YAxis allowDecimals={false} tick={{ fill: "#9ca3af", fontSize: 12 }} />
                            <Tooltip contentStyle={{ backgroundColor: "#111", border: "1px solid #333" }} labelStyle={{ color: "#eee" }} />
                            <Line type="monotone" dataKey="count" stroke="#f472b6" strokeWidth={2} dot={false} />
                          </LineChart>
                        </ResponsiveContainer>
                      </div>
                      <div className="mt-4 max-h-56 overflow-y-auto">
                        {logins.slice(0, 50).map((l) => (
                          <div key={l.id} className="flex items-center justify-between border-b border-neutral-800 py-2 text-sm">
                            <div className="text-gray-300">
                              <span className="font-semibold">{l.destination}</span> • {l.login_type}
                            </div>
                            <div className="text-gray-400">{new Date(l.created_at).toLocaleString()}</div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </Card>

                <Card className="mt-4 p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <FontAwesomeIcon icon={faChartLine} className="h-5 w-5" />
                      <h3 className="ml-4">Discord Activity (Last 90 Days)</h3>
                    </div>
                  </div>
                  <Separator className="my-2" />
                  {activitiesLoading ? (
                    <div className="flex w-full justify-center p-4">
                      <Loader2 className="animate-spin" />
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
                      <div className="h-64 w-full">
                        <div className="mb-2 flex items-center text-sm text-gray-300">
                          <FontAwesomeIcon icon={faMessage} className="mr-2" /> Messages per day
                        </div>
                        <ResponsiveContainer width="100%" height="100%">
                          <LineChart data={messageSeries} margin={{ left: 8, right: 16, top: 8, bottom: 8 }}>
                            <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                            <XAxis dataKey="date" tick={{ fill: "#9ca3af", fontSize: 12 }} />
                            <YAxis allowDecimals={false} tick={{ fill: "#9ca3af", fontSize: 12 }} />
                            <Tooltip contentStyle={{ backgroundColor: "#111", border: "1px solid #333" }} labelStyle={{ color: "#eee" }} />
                            <Line type="monotone" dataKey="count" stroke="#60a5fa" strokeWidth={2} dot={false} />
                          </LineChart>
                        </ResponsiveContainer>
                      </div>
                      <div className="h-64 w-full">
                        <div className="mb-2 flex items-center text-sm text-gray-300">
                          <FontAwesomeIcon icon={faFaceSmile} className="mr-2" /> Reactions per day
                        </div>
                        <ResponsiveContainer width="100%" height="100%">
                          <LineChart data={reactionSeries} margin={{ left: 8, right: 16, top: 8, bottom: 8 }}>
                            <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                            <XAxis dataKey="date" tick={{ fill: "#9ca3af", fontSize: 12 }} />
                            <YAxis allowDecimals={false} tick={{ fill: "#9ca3af", fontSize: 12 }} />
                            <Tooltip contentStyle={{ backgroundColor: "#111", border: "1px solid #333" }} labelStyle={{ color: "#eee" }} />
                            <Line type="monotone" dataKey="count" stroke="#34d399" strokeWidth={2} dot={false} />
                          </LineChart>
                        </ResponsiveContainer>
                      </div>
                    </div>
                  )}
                </Card>
              </div>
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default UserProfilePage;

