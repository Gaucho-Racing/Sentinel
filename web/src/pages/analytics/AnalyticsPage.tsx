import React from "react";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import { SENTINEL_API_URL } from "@/consts/config";
import { checkCredentials } from "@/lib/auth";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { notify } from "@/lib/notify";
import { User } from "@/models/user";
import Footer from "@/components/Footer";
import { AuthLoading } from "@/components/AuthLoading";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faArrowLeft,
  faChartBar,
  faChartLine,
  faChartPie,
} from "@fortawesome/free-solid-svg-icons";
import { Loader2 } from "lucide-react";
import {
  PieChart,
  Pie,
  Cell,
  Tooltip as ReTooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  LineChart,
  Line,
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

const COLORS = [
  "#f472b6",
  "#60a5fa",
  "#34d399",
  "#f59e0b",
  "#a78bfa",
  "#f87171",
  "#22d3ee",
  "#e879f9",
]; // on-brand

function AnalyticsPage() {
  const navigate = useNavigate();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);
  const [usersLoading, setUsersLoading] = React.useState(false);
  const [loginsLoading, setLoginsLoading] = React.useState(false);
  const [users, setUsers] = React.useState<User[]>([]);
  const [logins, setLogins] = React.useState<UserLogin[]>([]);

  React.useEffect(() => {
    checkAuth().then(() => {
      getUsers();
      getLogins();
    });
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

  const getUsers = async () => {
    setUsersLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) setUsers(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    setUsersLoading(false);
  };

  const getLogins = async () => {
    setLoginsLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/logins`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) setLogins(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    setLoginsLoading(false);
  };

  const genderData = React.useMemo(() => {
    const map: Record<string, number> = {};
    users.forEach((u) => {
      const g = (u.gender || "Unknown").trim() || "Unknown";
      map[g] = (map[g] || 0) + 1;
    });
    return Object.entries(map).map(([name, value]) => ({ name, value }));
  }, [users]);

  const gradYearData = React.useMemo(() => {
    const map: Record<string, number> = {};
    users.forEach((u) => {
      const y = u.graduation_year ? u.graduation_year.toString() : "Unknown";
      map[y] = (map[y] || 0) + 1;
    });
    return Object.entries(map)
      .sort((a, b) =>
        a[0] === "Unknown"
          ? 1
          : b[0] === "Unknown"
            ? -1
            : parseInt(a[0]) - parseInt(b[0]),
      )
      .map(([year, count]) => ({ year, count }));
  }, [users]);

  const subteamData = React.useMemo(() => {
    const map: Record<string, number> = {};
    users.forEach((u) =>
      u.subteams.forEach((s) => (map[s.name] = (map[s.name] || 0) + 1)),
    );
    return Object.entries(map)
      .sort((a, b) => b[1] - a[1])
      .map(([name, count]) => ({ name, count }));
  }, [users]);

  const rolesData = React.useMemo(() => {
    const map: Record<string, number> = {};
    users.forEach((u) =>
      u.roles
        .filter((r) => r.startsWith("d_"))
        .forEach((r) => (map[r] = (map[r] || 0) + 1)),
    );
    return Object.entries(map).map(([name, value]) => ({ name, value }));
  }, [users]);

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

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col justify-between">
          <div className="p-4 lg:p-32 lg:pt-16">
            <div className="mb-4">
              <Button
                variant={"ghost"}
                onClick={() => navigate("/")}
                className="flex items-center"
              >
                <FontAwesomeIcon
                  icon={faArrowLeft}
                  className="mr-2 h-4 w-4 text-gray-400"
                />
                Back to home
              </Button>
            </div>
            <h1>Analytics</h1>
            {(usersLoading || loginsLoading) && (
              <div className="mt-4 flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin" />
              </div>
            )}
            <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-2">
              <Card className="p-4">
                <div className="flex items-center">
                  <FontAwesomeIcon icon={faChartPie} className="h-5 w-5" />
                  <h3 className="ml-4">Gender Distribution</h3>
                </div>
                <Separator className="my-2" />
                <div className="h-64 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={genderData}
                        dataKey="value"
                        nameKey="name"
                        outerRadius={100}
                      >
                        {genderData.map((_, index) => (
                          <Cell
                            key={index}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <ReTooltip
                        contentStyle={{
                          backgroundColor: "#111",
                          border: "1px solid #333",
                        }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              <Card className="p-4">
                <div className="flex items-center">
                  <FontAwesomeIcon icon={faChartBar} className="h-5 w-5" />
                  <h3 className="ml-4">Graduation Year</h3>
                </div>
                <Separator className="my-2" />
                <div className="h-64 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart
                      data={gradYearData}
                      margin={{ left: 8, right: 16, top: 8, bottom: 8 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                      <XAxis
                        dataKey="year"
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                      />
                      <YAxis
                        allowDecimals={false}
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                      />
                      <ReTooltip
                        contentStyle={{
                          backgroundColor: "#111",
                          border: "1px solid #333",
                        }}
                      />
                      <Bar dataKey="count" fill="#60a5fa" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              <Card className="p-4">
                <div className="flex items-center">
                  <FontAwesomeIcon icon={faChartBar} className="h-5 w-5" />
                  <h3 className="ml-4">Subteam Sizes</h3>
                </div>
                <Separator className="my-2" />
                <div className="h-64 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart
                      data={subteamData}
                      margin={{ left: 8, right: 16, top: 8, bottom: 80 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                      <XAxis
                        dataKey="name"
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                        angle={-35}
                        textAnchor="end"
                        interval={0}
                        height={60}
                      />
                      <YAxis
                        allowDecimals={false}
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                      />
                      <ReTooltip
                        contentStyle={{
                          backgroundColor: "#111",
                          border: "1px solid #333",
                        }}
                      />
                      <Bar dataKey="count" fill="#34d399" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              <Card className="p-4">
                <div className="flex items-center">
                  <FontAwesomeIcon icon={faChartLine} className="h-5 w-5" />
                  <h3 className="ml-4">Logins (Last 30 Days)</h3>
                </div>
                <Separator className="my-2" />
                <div className="h-64 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart
                      data={loginSeries}
                      margin={{ left: 8, right: 16, top: 8, bottom: 8 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2a" />
                      <XAxis
                        dataKey="date"
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                      />
                      <YAxis
                        allowDecimals={false}
                        tick={{ fill: "#9ca3af", fontSize: 12 }}
                      />
                      <ReTooltip
                        contentStyle={{
                          backgroundColor: "#111",
                          border: "1px solid #333",
                        }}
                      />
                      <Line
                        type="monotone"
                        dataKey="count"
                        stroke="#f472b6"
                        strokeWidth={2}
                        dot={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              <Card className="p-4">
                <div className="flex items-center">
                  <FontAwesomeIcon icon={faChartPie} className="h-5 w-5" />
                  <h3 className="ml-4">Roles</h3>
                </div>
                <Separator className="my-2" />
                <div className="h-64 w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={rolesData}
                        dataKey="value"
                        nameKey="name"
                        outerRadius={100}
                      >
                        {rolesData.map((_, index) => (
                          <Cell
                            key={index}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <ReTooltip
                        contentStyle={{
                          backgroundColor: "#111",
                          border: "1px solid #333",
                        }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default AnalyticsPage;
