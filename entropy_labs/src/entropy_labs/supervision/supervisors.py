from typing import Callable, Any
from .config import SupervisionDecision, SupervisionDecisionType

def divide_supervisor(min_a: int=0, min_b: int = 0):
    """
    Decorator that creates a supervisor function to check if both numbers are above specified minimums.
    
    Args:
        min_a (int): Minimum allowed value for the first argument. Default is 0.
        min_b (int): Minimum allowed value for the second argument. Default is 0.
    
    Returns:
        Callable: A supervisor function that checks the arguments against the specified minimums.
    """
    def supervisor(func: Callable, a: float, b: float) -> SupervisionDecision:
        """Supervisor that checks if both numbers are above the specified minimums."""
        if a > min_a and b > min_b:
            return SupervisionDecision(
                decision=SupervisionDecisionType.APPROVE,
                explanation=f"Both numbers are above their minimums (a > {min_a}, b > {min_b})."
            )
        else:
            return SupervisionDecision(
                decision=SupervisionDecisionType.ESCALATE,
                explanation=f"One or both numbers are below their minimums (a <= {min_a} or b <= {min_b})."
            )
    
    supervisor.__name__ = divide_supervisor.__name__
    return supervisor
