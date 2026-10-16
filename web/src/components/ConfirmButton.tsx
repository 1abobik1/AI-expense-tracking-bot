import { useEffect, useState, type ReactNode } from "react";

// Подтверждение вторым нажатием: window.confirm не работает во встроенном браузере Telegram.
export default function ConfirmButton({
  onConfirm,
  children,
}: {
  onConfirm: () => void;
  children: ReactNode;
}) {
  const [armed, setArmed] = useState(false);
  useEffect(() => {
    if (!armed) return;
    const t = setTimeout(() => setArmed(false), 4000);
    return () => clearTimeout(t);
  }, [armed]);

  return (
    <button
      className={armed ? "danger armed" : "danger"}
      onClick={() => (armed ? (setArmed(false), onConfirm()) : setArmed(true))}
    >
      {armed ? "Точно?" : children}
    </button>
  );
}
