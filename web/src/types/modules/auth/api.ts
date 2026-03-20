import type { User } from "../users";
import { z } from "zod";
import type { authenticateSchema, registerSchema } from "./schema";



export type RegisterUserPayload = z.infer<typeof registerSchema>;

export type AuthenticatePayload = z.infer<typeof authenticateSchema>;

export type UserWithToken = {
  user: User;
  token: string;
};
