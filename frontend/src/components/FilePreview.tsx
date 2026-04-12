import { useTranslation } from 'react-i18next';

interface FilePreviewProps {
    url: string;
    mimeType: string;
    title?: string;
    className?: string;
}

export function FilePreview({ url, mimeType, title, className = "w-full h-full rounded-xl border border-gray-300 dark:border-gray-800 shadow-inner bg-white" }: FilePreviewProps) {
    const { t } = useTranslation();

    const isImage = mimeType.startsWith('image/');

    if (isImage) {
        return (
            <div className="flex items-center justify-center w-full h-full overflow-auto bg-gray-100 dark:bg-gray-900 rounded-xl">
                <img src={url} alt={title || t('common.preview')} className="max-w-full max-h-full object-contain shadow-lg" />
            </div>
        );
    }

    // Default to iframe for PDF or unknown
    return (
        <iframe 
            src={`${url}#toolbar=0`} 
            className={className} 
            title={title || t('common.preview')} 
        />
    );
}
