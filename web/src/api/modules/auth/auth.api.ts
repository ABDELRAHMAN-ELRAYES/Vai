import type {
  RegisterUserPayload,
  UserWithToken,
  AuthenticatePayload,
} from "@/types/modules/auth/api";

import { apiClient } from "../../client";
import type { User } from "@/types/modules/users";

const ROUTER_PREFIX = "/auth";

export const authApi = {
  /*
   * Register a new user with entered dataa
   */
  register(payload: RegisterUserPayload) {
    return apiClient.post<{data: UserWithToken}>(`${ROUTER_PREFIX}/register`, {
      first_name: payload.firstName,
      last_name: payload.lastName,
      email: payload.email,
      password: payload.password,
    });
  },
  /*
   * Activate the registered user
   */
  activate(token: string) {
    return apiClient.post(`${ROUTER_PREFIX}/activate/${token}`);
  },
  /*
   *   Authenticate a user
   */
  login(payload: AuthenticatePayload) {
    return apiClient.post<{data: UserWithToken}>(`${ROUTER_PREFIX}/login`, payload);
  },
  /*
   *   Logout a user
   */
  logout() {
    return apiClient.post(`${ROUTER_PREFIX}/logout`);
  },
  /*
   *   Get current user
   */
  getMe() {
    return apiClient.get<{data: User}>(`${ROUTER_PREFIX}/me`);
  },
};
