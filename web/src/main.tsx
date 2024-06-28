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

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
  },
  {
    path: "/auth/login/discord",
    element: <LoginDiscordPage />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
    <Toaster />
  </React.StrictMode>,
);
