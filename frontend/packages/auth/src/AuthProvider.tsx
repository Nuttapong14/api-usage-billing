'use client';

import React, { createContext, useContext, useEffect, useState, useCallback, useMemo, ReactNode } from 'react';
import { KeycloakAdapter } from './keycloak-adapter';
import { AuthContextValue, AuthError, AuthState, AuthTokens, KeycloakConfig, LoginOptions, LogoutOptions, User } from './types';

const AuthContext = createContext<AuthContextValue | null>(null);

interface AuthProviderProps {
  children: ReactNode;
  config: KeycloakConfig;
  onAuthenticated?: (user: User, tokens: AuthTokens) => void;
  onLogout?: () => void;
  onError?: (error: AuthError) => void;
  LoadingComponent?: React.ComponentType;
}

export function AuthProvider({
  children,
  config,
  onAuthenticated,
  onLogout,
  onError,
  LoadingComponent,
}: AuthProviderProps) {
  const [adapter] = useState(() => new KeycloakAdapter(config));
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    user: null,
    tokens: null,
    error: null,
  });

  useEffect(() => {
    const initAuth = async () => {
      try {
        adapter.setOnTokenRefresh((tokens) => {
          setState((prev) => ({ ...prev, tokens }));
        });

        adapter.setOnAuthError((error) => {
          const authError: AuthError = {
            code: 'token_refresh_failed',
            message: error.message,
          };
          setState((prev) => ({ ...prev, error: authError }));
          onError?.(authError);
        });

        const authenticated = await adapter.init();

        if (authenticated) {
          const user = adapter.getUser();
          const tokens = adapter.getTokens();

          setState({
            isAuthenticated: true,
            isLoading: false,
            user,
            tokens,
            error: null,
          });

          if (user && tokens) {
            onAuthenticated?.(user, tokens);
          }
        } else {
          setState({
            isAuthenticated: false,
            isLoading: false,
            user: null,
            tokens: null,
            error: null,
          });
        }
      } catch (error) {
        const authError: AuthError = {
          code: 'init_failed',
          message: error instanceof Error ? error.message : 'Authentication initialization failed',
        };

        setState({
          isAuthenticated: false,
          isLoading: false,
          user: null,
          tokens: null,
          error: authError,
        });

        onError?.(authError);
      }
    };

    initAuth();

    return () => {
      adapter.destroy();
    };
  }, [adapter, onAuthenticated, onError]);

  const login = useCallback(
    async (options?: LoginOptions) => {
      try {
        await adapter.login(options);
      } catch (error) {
        const authError: AuthError = {
          code: 'login_failed',
          message: error instanceof Error ? error.message : 'Login failed',
        };
        setState((prev) => ({ ...prev, error: authError }));
        onError?.(authError);
      }
    },
    [adapter, onError]
  );

  const logout = useCallback(
    async (options?: LogoutOptions) => {
      try {
        await adapter.logout(options);
        setState({
          isAuthenticated: false,
          isLoading: false,
          user: null,
          tokens: null,
          error: null,
        });
        onLogout?.();
      } catch (error) {
        const authError: AuthError = {
          code: 'logout_failed',
          message: error instanceof Error ? error.message : 'Logout failed',
        };
        setState((prev) => ({ ...prev, error: authError }));
        onError?.(authError);
      }
    },
    [adapter, onLogout, onError]
  );

  const refreshToken = useCallback(async () => {
    return adapter.refreshToken();
  }, [adapter]);

  const hasRole = useCallback(
    (role: string) => {
      return adapter.hasRole(role);
    },
    [adapter]
  );

  const hasPermission = useCallback(
    (permission: string) => {
      return adapter.hasPermission(permission);
    },
    [adapter]
  );

  const hasAnyRole = useCallback(
    (roles: string[]) => {
      return roles.some((role) => adapter.hasRole(role));
    },
    [adapter]
  );

  const hasAllRoles = useCallback(
    (roles: string[]) => {
      return roles.every((role) => adapter.hasRole(role));
    },
    [adapter]
  );

  const getAccessToken = useCallback(() => {
    return adapter.getAccessToken();
  }, [adapter]);

  const updateProfile = useCallback(
    async (data: Partial<User>) => {
      // Profile update would be handled via API
      setState((prev) => ({
        ...prev,
        user: prev.user ? { ...prev.user, ...data } : null,
      }));
    },
    []
  );

  const value = useMemo<AuthContextValue>(
    () => ({
      ...state,
      login,
      logout,
      refreshToken,
      hasRole,
      hasPermission,
      hasAnyRole,
      hasAllRoles,
      getAccessToken,
      updateProfile,
    }),
    [state, login, logout, refreshToken, hasRole, hasPermission, hasAnyRole, hasAllRoles, getAccessToken, updateProfile]
  );

  if (state.isLoading) {
    if (LoadingComponent) {
      return <LoadingComponent />;
    }
    return null;
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function useUser(): User | null {
  const { user } = useAuth();
  return user;
}

export function useIsAuthenticated(): boolean {
  const { isAuthenticated } = useAuth();
  return isAuthenticated;
}
