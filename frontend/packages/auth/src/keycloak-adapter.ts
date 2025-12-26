import Keycloak, { KeycloakInitOptions, KeycloakLoginOptions, KeycloakLogoutOptions } from 'keycloak-js';
import { AuthTokens, KeycloakConfig, LoginOptions, LogoutOptions, User } from './types';

export class KeycloakAdapter {
  private keycloak: Keycloak;
  private config: KeycloakConfig;
  private refreshIntervalId: number | null = null;
  private onTokenRefresh?: (tokens: AuthTokens) => void;
  private onAuthError?: (error: Error) => void;

  constructor(config: KeycloakConfig) {
    this.config = config;
    this.keycloak = new Keycloak({
      url: config.url,
      realm: config.realm,
      clientId: config.clientId,
    });
  }

  async init(): Promise<boolean> {
    const initOptions: KeycloakInitOptions = {
      onLoad: 'check-sso',
      silentCheckSsoRedirectUri: this.config.silentCheckSsoRedirectUri,
      pkceMethod: 'S256',
      checkLoginIframe: false,
      enableLogging: this.config.enableLogging,
    };

    try {
      const authenticated = await this.keycloak.init(initOptions);

      if (authenticated) {
        this.setupTokenRefresh();
      }

      return authenticated;
    } catch (error) {
      console.error('Keycloak initialization failed:', error);
      throw error;
    }
  }

  async login(options?: LoginOptions): Promise<void> {
    const loginOptions: KeycloakLoginOptions = {
      redirectUri: options?.redirectUri || this.config.redirectUri || window.location.origin,
      prompt: options?.prompt,
      loginHint: options?.loginHint,
      idpHint: options?.idpHint,
      scope: options?.scope,
    };

    await this.keycloak.login(loginOptions);
  }

  async logout(options?: LogoutOptions): Promise<void> {
    this.stopTokenRefresh();

    const logoutOptions: KeycloakLogoutOptions = {
      redirectUri: options?.redirectUri || window.location.origin,
    };

    await this.keycloak.logout(logoutOptions);
  }

  async refreshToken(): Promise<boolean> {
    try {
      const refreshed = await this.keycloak.updateToken(30);

      if (refreshed && this.onTokenRefresh) {
        this.onTokenRefresh(this.getTokens()!);
      }

      return refreshed;
    } catch (error) {
      console.error('Token refresh failed:', error);
      if (this.onAuthError) {
        this.onAuthError(error as Error);
      }
      return false;
    }
  }

  isAuthenticated(): boolean {
    return this.keycloak.authenticated || false;
  }

  getTokens(): AuthTokens | null {
    if (!this.keycloak.token) {
      return null;
    }

    return {
      accessToken: this.keycloak.token,
      refreshToken: this.keycloak.refreshToken || '',
      idToken: this.keycloak.idToken,
      expiresAt: (this.keycloak.tokenParsed?.exp || 0) * 1000,
    };
  }

  getUser(): User | null {
    if (!this.keycloak.tokenParsed) {
      return null;
    }

    const tokenParsed = this.keycloak.tokenParsed as Record<string, unknown>;
    const idTokenParsed = this.keycloak.idTokenParsed as Record<string, unknown> | undefined;

    // Extract roles from realm_access and resource_access
    const realmRoles = (tokenParsed.realm_access as { roles?: string[] })?.roles || [];
    const resourceRoles = Object.values(
      (tokenParsed.resource_access as Record<string, { roles?: string[] }>) || {}
    ).flatMap((r) => r.roles || []);

    return {
      id: tokenParsed.sub as string,
      email: (tokenParsed.email || idTokenParsed?.email) as string,
      name: (tokenParsed.name || idTokenParsed?.name || tokenParsed.preferred_username) as string,
      firstName: (tokenParsed.given_name || idTokenParsed?.given_name) as string | undefined,
      lastName: (tokenParsed.family_name || idTokenParsed?.family_name) as string | undefined,
      emailVerified: (tokenParsed.email_verified || false) as boolean,
      roles: [...new Set([...realmRoles, ...resourceRoles])],
      permissions: (tokenParsed.permissions as string[]) || [],
      organizationId: tokenParsed.organization_id as string | undefined,
      customerId: tokenParsed.customer_id as string | undefined,
      avatar: tokenParsed.picture as string | undefined,
    };
  }

  hasRole(role: string): boolean {
    return this.keycloak.hasRealmRole(role) || this.keycloak.hasResourceRole(role);
  }

  hasPermission(permission: string): boolean {
    const user = this.getUser();
    return user?.permissions.includes(permission) || false;
  }

  getAccessToken(): string | null {
    return this.keycloak.token || null;
  }

  setOnTokenRefresh(callback: (tokens: AuthTokens) => void): void {
    this.onTokenRefresh = callback;
  }

  setOnAuthError(callback: (error: Error) => void): void {
    this.onAuthError = callback;
  }

  private setupTokenRefresh(): void {
    // Refresh token 60 seconds before expiry
    const refreshInterval = 60 * 1000; // Check every minute

    this.refreshIntervalId = window.setInterval(() => {
      if (this.keycloak.isTokenExpired(60)) {
        this.refreshToken();
      }
    }, refreshInterval);

    // Also set up Keycloak's built-in token refresh handling
    this.keycloak.onTokenExpired = () => {
      this.refreshToken();
    };
  }

  private stopTokenRefresh(): void {
    if (this.refreshIntervalId) {
      clearInterval(this.refreshIntervalId);
      this.refreshIntervalId = null;
    }
  }

  destroy(): void {
    this.stopTokenRefresh();
  }
}
