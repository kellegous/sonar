import { Interceptor } from "@connectrpc/connect";

export const withSource: (source: string) => Interceptor = (source: string) => {
  return (next) => {
    return async (req) => {
      req.header.set("X-Rpc-Source", source);
      return next(req);
    };
  };
};
