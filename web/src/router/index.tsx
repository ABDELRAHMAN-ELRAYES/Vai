import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { lazy, Suspense } from "react";
import { PATHS } from "./paths";

import RootLayout from "@/layouts/RootLayout";
import NotFound from "@/pages/NotFount";
import { Loader } from "@/components/loaders/loader";

// Lazy loaded pages
const Home = lazy(() => import("@/pages/Home"));


// Susupense wrapper
function S({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<Loader />}>{children}</Suspense>;
}

const router = createBrowserRouter([
  {
    path: PATHS.HOME,
    element: <RootLayout />,
    errorElement: <NotFound />,
    children: [
      {
        index: true,
        element: (
          <S>
            <Home />
          </S>
        ),
      },
    ],
  },
  // Catch all 404
  {
    path: PATHS.NOT_FOUND,
    element: <NotFound />,
  },
]);

export default function AppRouter() {
  return <RouterProvider router={router} />;
}
