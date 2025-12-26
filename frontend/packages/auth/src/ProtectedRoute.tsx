'use client';

import React, { ReactNode } from 'react';
import { useAuth } from './AuthProvider';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredRoles?: string[];
  requiredPermissions?: string[];
  requireAll?: boolean;
  FallbackComponent?: React.ComponentType;
  UnauthorizedComponent?: React.ComponentType;
  LoadingComponent?: React.ComponentType;
}

export function ProtectedRoute({
  children,
  requiredRoles = [],
  requiredPermissions = [],
  requireAll = true,
  FallbackComponent,
  UnauthorizedComponent,
  LoadingComponent,
}: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, hasRole, hasPermission, login } = useAuth();

  if (isLoading) {
    if (LoadingComponent) {
      return <LoadingComponent />;
    }
    return null;
  }

  if (!isAuthenticated) {
    if (FallbackComponent) {
      return <FallbackComponent />;
    }
    // Auto-redirect to login
    login();
    return null;
  }

  // Check role requirements
  if (requiredRoles.length > 0) {
    const hasRequiredRoles = requireAll
      ? requiredRoles.every((role) => hasRole(role))
      : requiredRoles.some((role) => hasRole(role));

    if (!hasRequiredRoles) {
      if (UnauthorizedComponent) {
        return <UnauthorizedComponent />;
      }
      return <DefaultUnauthorized />;
    }
  }

  // Check permission requirements
  if (requiredPermissions.length > 0) {
    const hasRequiredPermissions = requireAll
      ? requiredPermissions.every((perm) => hasPermission(perm))
      : requiredPermissions.some((perm) => hasPermission(perm));

    if (!hasRequiredPermissions) {
      if (UnauthorizedComponent) {
        return <UnauthorizedComponent />;
      }
      return <DefaultUnauthorized />;
    }
  }

  return <>{children}</>;
}

function DefaultUnauthorized() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900">Access Denied</h1>
        <p className="mt-2 text-gray-600">You don't have permission to access this page.</p>
      </div>
    </div>
  );
}

interface RequireAuthProps {
  children: ReactNode;
  fallback?: ReactNode;
}

export function RequireAuth({ children, fallback }: RequireAuthProps) {
  const { isAuthenticated, isLoading, login } = useAuth();

  if (isLoading) {
    return fallback ?? null;
  }

  if (!isAuthenticated) {
    login();
    return fallback ?? null;
  }

  return <>{children}</>;
}

interface RequireRoleProps {
  children: ReactNode;
  role: string;
  fallback?: ReactNode;
}

export function RequireRole({ children, role, fallback }: RequireRoleProps) {
  const { hasRole } = useAuth();

  if (!hasRole(role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

interface RequirePermissionProps {
  children: ReactNode;
  permission: string;
  fallback?: ReactNode;
}

export function RequirePermission({ children, permission, fallback }: RequirePermissionProps) {
  const { hasPermission } = useAuth();

  if (!hasPermission(permission)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
