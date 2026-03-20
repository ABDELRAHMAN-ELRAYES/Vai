export interface RequestOptions extends Omit<RequestInit, "body"> {
  // append Query string params to the URL
  params?: Record<string, string | number | boolean | undefined>;
  // the request body
  body?: unknown;
}
