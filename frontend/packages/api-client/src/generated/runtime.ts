export function interpolatePath(
  template: string,
  params: Record<string, string | number | boolean>
): string {
  return template.replace(/\{([^}]+)\}/g, (_match, key) => {
    const value = params[key];
    if (value === undefined || value === null) {
      throw new Error(`Missing path parameter: ${key}`);
    }
    return encodeURIComponent(String(value));
  });
}
