export const PATHS = {
  HOME: "/",
  CONFIRM: "/confirm/:token",
  NOT_FOUND: "*",
} as const;

export type AppPath = (typeof PATHS)[keyof typeof PATHS];
