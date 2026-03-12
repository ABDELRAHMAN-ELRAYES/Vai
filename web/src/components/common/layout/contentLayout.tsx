import { Header } from "../heade";

export const ContentLayout = () => {
  return (
    <div className="flex flex-1 flex-col bg-white mx-2 my-2 border border-border rounded-lg min-w-0">
      <Header pageName={"Home"} />
      <div className="px-4 py-3">
        <p className="text-4xl text-black">Home</p>
      </div>
    </div>
  );
};
