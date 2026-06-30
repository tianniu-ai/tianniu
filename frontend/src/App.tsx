import {useCallback, useEffect, useState} from 'react'
import Sidebar from './components/Sidebar'
import ChatPanel from './components/ChatPanel'
import AssistantThread from './components/assistant-ui/thread'
import AssistantThreadList from './components/assistant-ui/thread-list'
import {BabyAgentRuntimeProvider} from './components/assistant-ui/runtime-provider'
import AuthModal from './components/AuthModal'
import {type ConversationVO, isLoggedIn, listConversations,} from './api'


export default function App() {
    return <AssistantUIApp/>
}

function AssistantUIApp() {
    const [showAuthModal, setShowAuthModal] = useState(!isLoggedIn())

    useEffect(() => {
        const handleAuthRequired = () => {
            setShowAuthModal(true)
        }
        window.addEventListener('auth_required', handleAuthRequired)
        return () => window.removeEventListener('auth_required', handleAuthRequired)
    }, [])

    const handleLoginSuccess = useCallback(() => {
        // Refresh page after login
        window.location.reload()
    }, [])

    return (
        <BabyAgentRuntimeProvider>
            <div style={{display: 'flex', width: '100%', height: '100', overflow: 'hidden'}}>
                <AssistantThreadList/>
                <AssistantThread/>
            </div>
            <AuthModal
                open={showAuthModal}
                onOpenChange={setShowAuthModal}
                onLoginSuccess={handleLoginSuccess}
            />
        </BabyAgentRuntimeProvider>
    )
}