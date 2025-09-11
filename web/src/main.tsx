import React from "react";
import ReactDOM from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import "./index.css";
import "/node_modules/react-grid-layout/css/styles.css";
import "/node_modules/react-resizable/css/styles.css";
import "mapbox-gl/dist/mapbox-gl.css";
import { Toaster } from "./components/ui/sonner.tsx";
import App from "./App.tsx";
import LoginDiscordPage from "@/pages/auth/LoginDiscordPage.tsx";
import LoginPage from "@/pages/auth/LoginPage.tsx";
import EditUserPage from "@/pages/users/EditUserPage.tsx";
import ApplicationsPage from "@/pages/applications/ApplicationsPage.tsx";
import AuthorizePage from "@/pages/oauth/AuthorizePage.tsx";
import UsersPage from "@/pages/users/UsersPage.tsx";
import UserProfilePage from "@/pages/users/UserProfilePage.tsx";
import AnalyticsPage from "@/pages/analytics/AnalyticsPage.tsx";

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
  },
  {
    path: "/auth/login",
    element: <LoginPage />,
  },
  {
    path: "/auth/login/discord",
    element: <LoginDiscordPage />,
  },
  {
    path: "/oauth/authorize",
    element: <AuthorizePage />,
  },
  {
    path: "/users",
    element: <UsersPage />,
  },
  {
    path: "/users/:id",
    element: <UserProfilePage />,
  },
  {
    path: "/users/:id/edit",
    element: <EditUserPage />,
  },
  {
    path: "/analytics",
    element: <AnalyticsPage />,
  },
  {
    path: "/applications",
    element: <ApplicationsPage />,
  },
  {
    path: "/applications/:id",
    element: <ApplicationsPage />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
    <Toaster />
  </React.StrictMode>,
);
