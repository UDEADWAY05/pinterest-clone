export const PageLoader = () => {
    return (
        // Фиксируем на весь экран
        <div className="fixed inset-0 flex items-center justify-center bg-white z-50">
            {/* Имитация логотипа Pinterest (красный круг, который пульсирует) */}
            <div className="w-16 h-16 bg-red-600 rounded-full animate-pulse flex items-center justify-center">
                <span className="text-white font-bold text-3xl font-serif">P</span>
            </div>
        </div>
    );
};