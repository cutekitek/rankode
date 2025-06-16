import { useMemo } from "react";
import { useAuth } from "./AuthContext";

import { AuthApi, TasksApi, Configuration, TopicsApi, TestCasesApi, AttemptsApi, UsersApi } from "../api/index";

export const useApi = () => {
  const {token} = useAuth();
  const apis = useMemo(() => {
    const config = new Configuration({
      basePath: import.meta.env.VITE_API_BASE_PATH || "http://localhost:4000/api",
      headers: token ? {
        "Authorization": `Bearer ${token}`
      }: undefined,
    });
    return {
      auth: new AuthApi(config),
      tasks: new TasksApi(config),
      topics: new TopicsApi(config),
      testCases: new TestCasesApi(config),
      attempts: new AttemptsApi(config),
      users: new UsersApi(config),
    };
  }, [token]);

  return apis;
};
