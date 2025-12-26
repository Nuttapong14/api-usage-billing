export interface User {
  id: string;
  email: string;
  name: string;
  firstName?: string;
  lastName?: string;
  emailVerified: boolean;
  roles: string[];
  permissions: string[];
  organizationId?: string;
  customerId?: string;
  avatar?: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  idToken?: string;
  expiresAt: number;
}

export interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: User | null;
  tokens: AuthTokens | null;
  error: AuthError | null;
}

export interface AuthError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface KeycloakConfig {
  url: string;
  realm: string;
  clientId: string;
  redirectUri?: string;
  silentCheckSsoRedirectUri?: string;
  enableLogging?: boolean;
}

export interface AuthContextValue extends AuthState {
  login: (options?: LoginOptions) => Promise<void>;
  logout: (options?: LogoutOptions) => Promise<void>;
  refreshToken: () => Promise<boolean>;
  hasRole: (role: string) => boolean;
  hasPermission: (permission: string) => boolean;
  hasAnyRole: (roles: string[]) => boolean;
  hasAllRoles: (roles: string[]) => boolean;
  getAccessToken: () => string | null;
  updateProfile: (data: Partial<User>) => Promise<void>;
}

export interface LoginOptions {
  redirectUri?: string;
  prompt?: 'none' | 'login' | 'consent';
  loginHint?: string;
  idpHint?: string;
  scope?: string;
}

export interface LogoutOptions {
  redirectUri?: string;
}

export interface SessionInfo {
  sessionId: string;
  sessionState: string;
  issuedAt: number;
  expiresAt: number;
}
