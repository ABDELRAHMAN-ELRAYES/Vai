export const PATHS = {
  HOME: "/",
  NOT_FOUND: "*",
} as const;

export type AppPath = (typeof PATHS)[keyof typeof PATHS];
