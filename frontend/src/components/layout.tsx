import type { FC } from "react";
import { Outlet } from "react-router";
import { Header } from "./header";

export const Layout: FC = () => <>
    <Header />
    <Outlet />
</>