import { useEffect, useRef } from "react";

const INTERVAL_MS = 15_000;

// Раз в 15 секунд, пока вкладка видна, и сразу при возврате на неё.
export function useAutoRefresh(refresh: () => void) {
  const ref = useRef(refresh);
  ref.current = refresh;

  useEffect(() => {
    const tick = () => document.visibilityState === "visible" && ref.current();
    const timer = setInterval(tick, INTERVAL_MS);
    document.addEventListener("visibilitychange", tick);
    return () => {
      clearInterval(timer);
      document.removeEventListener("visibilitychange", tick);
    };
  }, []);
}
