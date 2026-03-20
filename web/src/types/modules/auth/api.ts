import type { User } from "../users";

export type RegisterUserPayload = {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
};

export type AuthenticatePayload = {
  email: string;
  password: string;
};

export type UserWithToken = {
  user: User;
  token: string;
};
