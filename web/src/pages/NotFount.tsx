import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

const NotFound = () => {
  return (
    <div className="h-full w-full p-3 sm:p-4 flex justify-center items-center">
      <div className="h-full flex flex-col sm:items-center justify-between sm:justify-center w-full max-w-3xl lg:max-w-[55rem] gap-7">
        <div className="sm:hidden" />

        <div className="flex flex-col sm:items-center gap-6">
          <span className="text-[80px] sm:text-[120px] font-semibold leading-none tracking-tighter text-slate-200 text-shadow-sm select-none">
            404
          </span>

          <div className="text-left sm:text-center">
            <h1 className="text-5xl text-[#2d2016] tracking-tight mb-2">
              Page not found.
            </h1>
            <p className="text-[15px] text-gray-400 leading-relaxed max-w-sm">
              The page you're looking for doesn't exist or may have been moved.
            </p>
          </div>

          <Button
            onClick={() => window.history.back()}
            className="mt-2 flex items-center gap-2 px-4 h-10 rounded-lg bg-black text-white text-[14px] shadow-sm hover:bg-neutral-800 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Go back
          </Button>
        </div>

        <div className="hidden sm:flex items-start gap-2 text-sm text-muted-foreground">
          <span className="leading-5">
            Lost? You can return{" "}
            <Link
              to="/"
              className="text-muted-foreground underline"
              onClick={() => (window.location.href = "/")}
            >
              home
            </Link>{" "}
            or contact{" "}
            <Link to="/" className="text-muted-foreground underline">
              support
            </Link>
            .
          </span>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
